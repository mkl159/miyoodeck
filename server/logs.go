package main

import (
	"net/http"
	"os"
	"strconv"
	"strings"
)

const LogFile = "/tmp/webdeck.log"

// handleLogs returns the last N lines of the server log (default 200, max 1000).
// GET /api/logs?lines=N
func handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	n := 200
	if v := r.URL.Query().Get("lines"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			n = parsed
		}
	}
	if n > 1000 {
		n = 1000
	}

	data, err := os.ReadFile(LogFile)
	if err != nil {
		jsonOK(w, map[string]string{"log": ""})
		return
	}

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	jsonOK(w, map[string]string{"log": strings.Join(lines, "\n")})
}
