package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Config file extensions that are editable
var editableExtensions = map[string]bool{
	".json": true, ".cfg": true, ".txt": true,
	".conf": true, ".ini": true, ".sh": true,
}

func handleConfigList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	searchDirs := []string{
		SysDir + "/config",
		"/mnt/SDCARD/RetroArch/.retroarch",
		"/mnt/SDCARD/App",
	}

	type ConfigFile struct {
		Name    string `json:"name"`
		Path    string `json:"path"`
		Size    int64  `json:"size"`
		ModTime string `json:"mod_time"`
		Type    string `json:"type"`
	}

	var configs []ConfigFile
	for _, dir := range searchDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(e.Name()))
			if !editableExtensions[ext] {
				continue
			}
			info, _ := e.Info()
			var size int64
			var modTime string
			if info != nil {
				size = info.Size()
				modTime = info.ModTime().Format("2006-01-02 15:04")
			}
			configs = append(configs, ConfigFile{
				Name:    e.Name(),
				Path:    filepath.Join(dir, e.Name()),
				Size:    size,
				ModTime: modTime,
				Type:    strings.TrimPrefix(ext, "."),
			})
		}
	}

	jsonOK(w, configs)
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		readConfig(w, r)
	case http.MethodPost, http.MethodPut:
		writeConfig(w, r)
	default:
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func readConfig(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		jsonError(w, "Missing path", http.StatusBadRequest)
		return
	}
	path = filepath.Clean(path)
	if !strings.HasPrefix(path, SDCard) && !strings.HasPrefix(path, SysDir) {
		jsonError(w, "Access denied", http.StatusForbidden)
		return
	}

	data, err := os.ReadFile(path)
	if err != nil {
		jsonError(w, "Cannot read file: "+err.Error(), http.StatusNotFound)
		return
	}

	jsonOK(w, map[string]string{
		"path":    path,
		"content": string(data),
	})
}

func writeConfig(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	req.Path = filepath.Clean(req.Path)
	if !strings.HasPrefix(req.Path, SDCard) && !strings.HasPrefix(req.Path, SysDir) {
		jsonError(w, "Access denied", http.StatusForbidden)
		return
	}

	ext := strings.ToLower(filepath.Ext(req.Path))
	if !editableExtensions[ext] {
		jsonError(w, "File type not editable", http.StatusForbidden)
		return
	}

	// Create backup before writing
	backupPath := req.Path + ".bak." + time.Now().Format("20060102_150405")
	if existing, err := os.ReadFile(req.Path); err == nil {
		os.WriteFile(backupPath, existing, 0644)
	}

	if err := os.WriteFile(req.Path, []byte(req.Content), 0644); err != nil {
		jsonError(w, "Write failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]string{
		"message": "Saved",
		"backup":  backupPath,
		"path":    req.Path,
	})
}
