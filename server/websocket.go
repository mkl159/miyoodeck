package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins (home network assumption)
	},
}

// Hub manages all active WebSocket connections.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
}

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

var hub = &Hub{
	clients:    make(map[*Client]bool),
	broadcast:  make(chan []byte, 128),
	register:   make(chan *Client, 16),
	unregister: make(chan *Client, 16),
}

func (h *Hub) run() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("hub.run panic recovered: %v — restarting", r)
			go h.run()
		}
	}()
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Client is too slow — disconnect it cleanly
					delete(h.clients, client)
					close(client.send)
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) clientCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.clients)
}

// safeBroadcast sends to hub without blocking (drops the message if the
// broadcast channel is full to avoid deadlocking broadcastLoop).
func safeBroadcast(data []byte) {
	select {
	case hub.broadcast <- data:
	default:
		// channel full — drop this frame rather than stalling
	}
}

func handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WS upgrade error: %v", err)
		return
	}

	client := &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 64),
	}
	hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		if r := recover(); r != nil {
			log.Printf("writePump panic: %v", r)
		}
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("readPump panic: %v", r)
		}
		// Use non-blocking unregister to avoid deadlock if hub is stuck
		select {
		case hub.unregister <- c:
		default:
			// hub channel full — force-remove the client
			hub.mu.Lock()
			if _, ok := hub.clients[c]; ok {
				delete(hub.clients, c)
				close(c.send)
			}
			hub.mu.Unlock()
		}
		c.conn.Close()
	}()

	c.conn.SetReadLimit(4096)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
	}
}

type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// broadcastLoop sends system stats to all connected clients.
// CPU is sampled by a background goroutine (non-blocking here).
func broadcastLoop() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("broadcastLoop panic: %v — restarting", r)
			go broadcastLoop()
		}
	}()

	statsTicker := time.NewTicker(2 * time.Second)
	screenshotTicker := time.NewTicker(1500 * time.Millisecond)
	defer statsTicker.Stop()
	defer screenshotTicker.Stop()

	for {
		select {
		case <-statsTicker.C:
			if hub.clientCount() == 0 {
				continue
			}
			info := SystemInfo{
				CPU:     readCPUUsage(),
				RAM:     readRAM(),
				Battery: readBattery(),
				IP:      getLocalIP(),
				Uptime:  readUptime(),
				CPUFreq: readCPUFreq(),
				Temp:    readCPUTemp(),
			}
			msg := WSMessage{Type: "stats", Data: map[string]interface{}{
				"cpu_percent":  info.CPU,
				"ram":          info.RAM,
				"battery":      info.Battery,
				"ip":           info.IP,
				"uptime":       info.Uptime,
				"cpu_freq_mhz": info.CPUFreq,
				"temp_c":       info.Temp,
				"game_running": isGameRunning(),
			}}
			if data, err := json.Marshal(msg); err == nil {
				safeBroadcast(data)
			}

		case <-screenshotTicker.C:
			if hub.clientCount() == 0 {
				continue
			}
			if isGameRunning() {
				screenshotTicker.Reset(3 * time.Second)
			} else {
				screenshotTicker.Reset(1500 * time.Millisecond)
			}
			b64 := screenshotBase64()
			if b64 == "" {
				continue
			}
			msg := WSMessage{Type: "screenshot", Data: b64}
			if data, err := json.Marshal(msg); err == nil {
				safeBroadcast(data)
			}
		}
	}
}
