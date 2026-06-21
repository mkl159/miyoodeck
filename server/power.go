package main

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"time"
)

// handlePower reboots or powers off the console.
//
// POST /api/system/power  {"action": "reboot" | "poweroff"}
//
// The command is issued from a short-lived goroutine so the HTTP response is
// flushed to the client *before* the system goes down. A `sync` is run first
// to flush pending writes (e.g. in-game saves) to the SD card.
//
// Note: a true "sleep/suspend" is intentionally not exposed — suspending the
// Miyoo drops WiFi and freezes this server, which would make the dashboard
// unreachable and leave no way to wake the device remotely.
func handlePower(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Action string `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var shellCmd, msg string
	switch req.Action {
	case "reboot":
		shellCmd = "sync; reboot"
		msg = "Rebooting…"
	case "poweroff", "shutdown":
		shellCmd = "sync; poweroff"
		msg = "Powering off…"
	default:
		jsonError(w, "Unknown action: "+req.Action, http.StatusBadRequest)
		return
	}

	jsonOK(w, map[string]string{"message": msg})

	go func() {
		// Give the response time to reach the browser before going down.
		time.Sleep(500 * time.Millisecond)
		exec.Command("sh", "-c", shellCmd).Run()
	}()
}
