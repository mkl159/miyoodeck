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
	FbWidth   = 640
	FbHeight  = 480
)

// tokenStore holds valid session tokens (in-memory, cleared on restart)
var tokenStore = map[string]time.Time{}

func main() {
	// Show IP address on startup for easy access
	ip := getLocalIP()
	fmt.Printf("\n╔══════════════════════════════════════╗\n")
	fmt.Printf("║       ONION WEB DECK v1.0            ║\n")
	fmt.Printf("╠══════════════════════════════════════╣\n")
	fmt.Printf("║  http://%-28s  ║\n", fmt.Sprintf("%s:%d", ip, Port))
	fmt.Printf("║  http://%-28s  ║\n", fmt.Sprintf("onion.local:%d", Port))
	fmt.Printf("╚══════════════════════════════════════╝\n\n")

	mux := http.NewServeMux()

	// Auth
	mux.HandleFunc("/api/auth/login", handleLogin)
	mux.HandleFunc("/api/auth/logout", handleLogout)
	mux.HandleFunc("/api/auth/setup", handleSetupPin)
	mux.HandleFunc("/api/auth/status", handleAuthStatus)

	// Protected API routes
	mux.HandleFunc("/api/system", auth(handleSystem))
	mux.HandleFunc("/api/systems", auth(handleSystems))
	mux.HandleFunc("/api/roms", auth(handleRoms))
	mux.HandleFunc("/api/launch", auth(handleLaunch))
	mux.HandleFunc("/api/files", auth(handleFiles))
	mux.HandleFunc("/api/upload", auth(handleUpload))
	mux.HandleFunc("/api/unzip", auth(handleUnzip))
	mux.HandleFunc("/api/delete", auth(handleDelete))
	mux.HandleFunc("/api/saves/backup", auth(handleSavesBackup))
	mux.HandleFunc("/api/screenshot", auth(handleScreenshot))
	mux.HandleFunc("/api/config/list", auth(handleConfigList))
	mux.HandleFunc("/api/config", auth(handleConfig))
	mux.HandleFunc("/ws", auth(handleWS))

	// Serve static frontend files
	fs := http.FileServer(http.Dir(StaticDir))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// SPA fallback: serve index.html for unknown paths
		path := StaticDir + r.URL.Path
		if _, err := os.Stat(path); os.IsNotExist(err) {
			http.ServeFile(w, r, StaticDir+"/index.html")
			return
		}
		fs.ServeHTTP(w, r)
	})

	// Start WebSocket hub
	go hub.run()

	// Periodically broadcast system stats to all WS clients
	go broadcastLoop()

	handler := corsMiddleware(mux)

	log.Printf("WebDeck server listening on :%d", Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", Port), handler); err != nil {
		log.Fatal(err)
	}
}

// ─── Auth middleware ───────────────────────────────────────────────────────────

func auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip auth if no PIN is configured
		if !pinConfigured() {
			next(w, r)
			return
		}
		// Check cookie
		cookie, err := r.Cookie("webdeck_token")
		if err != nil || !validToken(cookie.Value) {
			// Check Authorization header as fallback
			header := r.Header.Get("Authorization")
			token := strings.TrimPrefix(header, "Bearer ")
			if !validToken(token) {
				jsonError(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
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
		// No PIN set — auto-login
		token := newToken()
		setTokenCookie(w, token)
		jsonOK(w, map[string]string{"token": token, "message": "No PIN configured"})
		return
	}

	stored, _ := os.ReadFile(PinFile)
	hashed := hashPin(req.Pin)
	if strings.TrimSpace(string(stored)) != hashed {
		jsonError(w, "Wrong PIN", http.StatusUnauthorized)
		return
	}

	token := newToken()
	setTokenCookie(w, token)
	jsonOK(w, map[string]string{"token": token})
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("webdeck_token")
	if err == nil {
		delete(tokenStore, cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: "webdeck_token", Value: "", MaxAge: -1})
	jsonOK(w, map[string]string{"message": "Logged out"})
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
	hashed := hashPin(req.Pin)
	if err := os.WriteFile(PinFile, []byte(hashed), 0600); err != nil {
		jsonError(w, "Failed to save PIN: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]string{"message": "PIN set successfully"})
}

func handleAuthStatus(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, map[string]interface{}{
		"pin_configured": pinConfigured(),
	})
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func pinConfigured() bool {
	_, err := os.Stat(PinFile)
	return err == nil
}

func hashPin(pin string) string {
	h := sha256.Sum256([]byte(pin))
	return hex.EncodeToString(h[:])
}

func newToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	token := hex.EncodeToString(b)
	tokenStore[token] = time.Now().Add(24 * time.Hour)
	return token
}

func validToken(token string) bool {
	if token == "" {
		return false
	}
	exp, ok := tokenStore[token]
	if !ok {
		return false
	}
	if time.Now().After(exp) {
		delete(tokenStore, token)
		return false
	}
	return true
}

func setTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "webdeck_token",
		Value:    token,
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
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
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "unknown"
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ip := ipNet.IP.String()
				// Prefer wlan0 addresses (192.168.x.x range)
				if strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") {
					return ip
				}
			}
		}
	}
	return "0.0.0.0"
}
