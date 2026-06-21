package main

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"
)

// Favorites are persisted as a JSON array on the SD card so they survive
// reboots and server restarts.
const FavFile = "/mnt/SDCARD/.tmp_update/config/webdeck_favorites.json"

type FavEntry struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	System string `json:"system"`
}

var favMu sync.Mutex

func loadFavs() []FavEntry {
	favMu.Lock()
	defer favMu.Unlock()
	return readFavsLocked()
}

func readFavsLocked() []FavEntry {
	data, err := os.ReadFile(FavFile)
	if err != nil {
		return []FavEntry{}
	}
	var favs []FavEntry
	if json.Unmarshal(data, &favs) != nil {
		return []FavEntry{}
	}
	return favs
}

func saveFavsLocked(favs []FavEntry) error {
	data, _ := json.MarshalIndent(favs, "", "  ")
	return os.WriteFile(FavFile, data, 0644)
}

// handleFavorites lists, adds or removes favorite ROMs.
//
//	GET  /api/favorites                                  -> [FavEntry]
//	POST /api/favorites {action:"add|remove|toggle", path, name, system}
func handleFavorites(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		jsonOK(w, loadFavs())

	case http.MethodPost:
		var req struct {
			Action string `json:"action"`
			Path   string `json:"path"`
			Name   string `json:"name"`
			System string `json:"system"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if req.Path == "" {
			jsonError(w, "Missing path", http.StatusBadRequest)
			return
		}

		favMu.Lock()
		favs := readFavsLocked()
		idx := -1
		for i, f := range favs {
			if f.Path == req.Path {
				idx = i
				break
			}
		}
		exists := idx >= 0

		add := req.Action == "add" || (req.Action == "toggle" && !exists)
		remove := req.Action == "remove" || (req.Action == "toggle" && exists)

		if add && !exists {
			favs = append(favs, FavEntry{Name: req.Name, Path: req.Path, System: req.System})
		} else if remove && exists {
			favs = append(favs[:idx], favs[idx+1:]...)
		}
		err := saveFavsLocked(favs)
		favMu.Unlock()

		if err != nil {
			jsonError(w, "Cannot save favorites: "+err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w, map[string]interface{}{"favorite": add, "count": len(favs)})

	default:
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
