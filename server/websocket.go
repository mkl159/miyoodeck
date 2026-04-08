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
	broadcast:  make(chan []byte, 64),
	register:   make(chan *Client),
	unregister: make(chan *Client),
}

func (h *Hub) run() {
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
					close(client.send)
					delete(h.clients, client)
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
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
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
		hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
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
// Screenshot is only sent if clients are connected (saves CPU).
func broadcastLoop() {
	statsTicker := time.NewTicker(2 * time.Second)
	screenshotTicker := time.NewTicker(3 * time.Second)
	defer statsTicker.Stop()
	defer screenshotTicker.Stop()

	for {
		select {
		case <-statsTicker.C:
			if hub.clientCount() == 0 {
				continue // No clients → sleep mode
			}
			info := SystemInfo{
				CPU:     readCPUUsage(),
				RAM:     readRAM(),
				Battery: readBattery(),
				IP:      getLocalIP(),
				Uptime:  readUptime(),
				CPUFreq: readCPUFreq(),
			}
			msg := WSMessage{Type: "stats", Data: info}
			data, _ := json.Marshal(msg)
			hub.broadcast <- data

		case <-screenshotTicker.C:
			if hub.clientCount() == 0 {
				continue
			}
			b64 := screenshotBase64()
			if b64 == "" {
				continue
			}
			msg := WSMessage{Type: "screenshot", Data: b64}
			data, _ := json.Marshal(msg)
			hub.broadcast <- data
		}
	}
}
