package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	Port      = 8080
	StaticDir = "./www"
	SDCard    = "/mnt/SDCARD"
	SysDir    = "/mnt/SDCARD/.tmp_update"
	RomsDir   = "/mnt/SDCARD/Roms"
	SavesDir  = "/mnt/SDCARD/Saves/CurrentProfile/saves"
	EmuDir    = "/mnt/SDCARD/Emu"
	PinFile   = "/mnt/SDCARD/.tmp_update/config/webdeck_pin.txt"
	FbDevice  = "/dev/fb0"
)

// ─── Thread-safe token store ──────────────────────────────────────────────────

type tokenStore struct {
	mu     sync.RWMutex
	tokens map[string]time.Time
}

func newTokenStore() *tokenStore {
	ts := &tokenStore{tokens: make(map[string]time.Time)}
	go ts.cleanupLoop()
	return ts
}

func (ts *tokenStore) set(token string, exp time.Time) {
	ts.mu.Lock()
	ts.tokens[token] = exp
	ts.mu.Unlock()
}

func (ts *tokenStore) valid(token string) bool {
	if token == "" {
		return false
	}
	ts.mu.RLock()
	exp, ok := ts.tokens[token]
	ts.mu.RUnlock()
	if !ok {
		return false
	}
	if time.Now().After(exp) {
		ts.mu.Lock()
		delete(ts.tokens, token)
		ts.mu.Unlock()
		return false
	}
	return true
}

func (ts *tokenStore) delete(token string) {
	ts.mu.Lock()
	delete(ts.tokens, token)
	ts.mu.Unlock()
}

// cleanupLoop purges expired tokens every 10 minutes (fix: memory leak)
func (ts *tokenStore) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		ts.mu.Lock()
		for t, exp := range ts.tokens {
			if now.After(exp) {
				delete(ts.tokens, t)
			}
		}
		ts.mu.Unlock()
	}
}

var sessions = newTokenStore()

func main() {
	ip := getLocalIP()
	fmt.Printf("\n╔══════════════════════════════════════╗\n")
	fmt.Printf("║        MIYOODECK  v1.6               ║\n")
	fmt.Printf("╠══════════════════════════════════════╣\n")
	fmt.Printf("║  http://%-28s  ║\n", fmt.Sprintf("%s:%d", ip, Port))
	fmt.Printf("║  http://%-28s  ║\n", fmt.Sprintf("miyoodeck.local:%d", Port))
	fmt.Printf("╚══════════════════════════════════════╝\n\n")

	mux := http.NewServeMux()

	// Auth (public)
	mux.HandleFunc("/api/auth/login", handleLogin)
	mux.HandleFunc("/api/auth/logout", handleLogout)
	mux.HandleFunc("/api/auth/setup", handleSetupPin)
	mux.HandleFunc("/api/auth/status", handleAuthStatus)

	// Protected API
	mux.HandleFunc("/api/system", auth(handleSystem))
	mux.HandleFunc("/api/systems", auth(handleSystems))
	mux.HandleFunc("/api/roms", auth(handleRoms))
	mux.HandleFunc("/api/launch", auth(handleLaunch))
	mux.HandleFunc("/api/files", auth(handleFiles))
	mux.HandleFunc("/api/upload", auth(handleUpload))
	mux.HandleFunc("/api/unzip", auth(handleUnzip))
	mux.HandleFunc("/api/delete", auth(handleDelete))
	mux.HandleFunc("/api/download", auth(handleDownload))
	mux.HandleFunc("/api/saves/backup", auth(handleSavesBackup))
	mux.HandleFunc("/api/screenshot", auth(handleScreenshot))
	mux.HandleFunc("/api/config/list", auth(handleConfigList))
	mux.HandleFunc("/api/config", auth(handleConfig))
	mux.HandleFunc("/api/input/press", auth(handleInputPress))
	mux.HandleFunc("/ws", auth(handleWS))

	// SPA fallback
	fs := http.FileServer(http.Dir(StaticDir))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := StaticDir + r.URL.Path
		if _, err := os.Stat(path); os.IsNotExist(err) {
			http.ServeFile(w, r, StaticDir+"/index.html")
			return
		}
		fs.ServeHTTP(w, r)
	})

	go hub.run()
	go broadcastLoop()
	go startMDNS(ip)

	log.Printf("MiyooDeck listening on :%d", Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", Port), corsMiddleware(mux)); err != nil {
		log.Fatal(err)
	}
}

// ─── Auth middleware ───────────────────────────────────────────────────────────

func auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !pinConfigured() {
			next(w, r)
			return
		}
		// Check cookie
		if c, err := r.Cookie("webdeck_token"); err == nil && sessions.valid(c.Value) {
			next(w, r)
			return
		}
		// Check Authorization header
		if h := r.Header.Get("Authorization"); sessions.valid(strings.TrimPrefix(h, "Bearer ")) {
			next(w, r)
			return
		}
		// Fix #2: also check ?token= query param (used by saves backup download link)
		if t := r.URL.Query().Get("token"); sessions.valid(t) {
			next(w, r)
			return
		}
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ─── Auth handlers ─────────────────────────────────────────────────────────────

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Pin string `json:"pin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if !pinConfigured() {
		token := newToken()
		setTokenCookie(w, token)
		jsonOK(w, map[string]string{"token": token})
		return
	}
	stored, _ := os.ReadFile(PinFile)
	if strings.TrimSpace(string(stored)) != hashPin(req.Pin) {
		jsonError(w, "Wrong PIN", http.StatusUnauthorized)
		return
	}
	token := newToken()
	setTokenCookie(w, token)
	jsonOK(w, map[string]string{"token": token})
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie("webdeck_token"); err == nil {
		sessions.delete(c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: "webdeck_token", Value: "", MaxAge: -1, Path: "/"})
	jsonOK(w, map[string]string{"message": "ok"})
}

func handleSetupPin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Pin string `json:"pin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if len(req.Pin) < 4 {
		jsonError(w, "PIN must be at least 4 digits", http.StatusBadRequest)
		return
	}
	if err := os.WriteFile(PinFile, []byte(hashPin(req.Pin)), 0600); err != nil {
		jsonError(w, "Failed to save PIN: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]string{"message": "PIN set"})
}

func handleAuthStatus(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, map[string]bool{"pin_configured": pinConfigured()})
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func pinConfigured() bool {
	_, err := os.Stat(PinFile)
	return err == nil
}

// withinSD reports whether a cleaned path is the SD card root or strictly
// inside it. Using a trailing "/" guard prevents a sibling like
// "/mnt/SDCARDevil" from passing a naive HasPrefix check.
func withinSD(path string) bool {
	return path == SDCard || strings.HasPrefix(path, SDCard+"/")
}

func hashPin(pin string) string {
	h := sha256.Sum256([]byte(pin))
	return hex.EncodeToString(h[:])
}

func newToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	token := hex.EncodeToString(b)
	sessions.set(token, time.Now().Add(24*time.Hour))
	return token
}

func setTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name: "webdeck_token", Value: token,
		Path: "/", MaxAge: 86400,
		HttpOnly: false, SameSite: http.SameSiteLaxMode,
	})
}

func jsonOK(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func getLocalIP() string {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ip4 := ipNet.IP.To4(); ip4 != nil {
				ip := ip4.String()
				if strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") || strings.HasPrefix(ip, "172.") {
					return ip
				}
			}
		}
	}
	return "0.0.0.0"
}
