package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// backlightDir returns the first backlight device exposed by the kernel,
// e.g. /sys/class/backlight/backlight. Empty string if none is present.
func backlightDir() string {
	base := "/sys/class/backlight"
	entries, err := os.ReadDir(base)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		return filepath.Join(base, e.Name())
	}
	return ""
}

func readIntFile(path string) (int, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	v, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, false
	}
	return v, true
}

// handleBrightness reports or sets the screen backlight level.
//
//	GET  /api/system/brightness            -> {supported, value, max}
//	POST /api/system/brightness {value:N}  -> clamps and writes N
func handleBrightness(w http.ResponseWriter, r *http.Request) {
	dir := backlightDir()
	if dir == "" {
		jsonOK(w, map[string]interface{}{"supported": false})
		return
	}
	maxVal, ok := readIntFile(filepath.Join(dir, "max_brightness"))
	if !ok || maxVal <= 0 {
		maxVal = 100
	}

	switch r.Method {
	case http.MethodGet:
		cur, _ := readIntFile(filepath.Join(dir, "brightness"))
		jsonOK(w, map[string]interface{}{"supported": true, "value": cur, "max": maxVal})

	case http.MethodPost, http.MethodPut:
		var req struct {
			Value int `json:"value"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if req.Value < 0 {
			req.Value = 0
		}
		if req.Value > maxVal {
			req.Value = maxVal
		}
		if err := os.WriteFile(filepath.Join(dir, "brightness"), []byte(strconv.Itoa(req.Value)), 0644); err != nil {
			jsonError(w, "Cannot set brightness: "+err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w, map[string]interface{}{"supported": true, "value": req.Value, "max": maxVal})

	default:
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
