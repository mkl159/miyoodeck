package main

import (
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// collectRoms walks every system folder under RomsDir and returns all ROMs,
// optionally filtered to a single system and/or a case-insensitive name query.
func collectRoms(systemFilter, query string) []ROM {
	query = strings.ToLower(strings.TrimSpace(query))

	systems, err := os.ReadDir(RomsDir)
	if err != nil {
		return nil
	}

	var roms []ROM
	for _, s := range systems {
		if !s.IsDir() || strings.HasPrefix(s.Name(), ".") {
			continue
		}
		if systemFilter != "" && !strings.EqualFold(s.Name(), systemFilter) {
			continue
		}
		sysPath := filepath.Join(RomsDir, s.Name())
		entries, err := os.ReadDir(sysPath)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(e.Name()))
			if !romExtensions[ext] {
				continue
			}
			name := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
			if query != "" && !strings.Contains(strings.ToLower(name), query) {
				continue
			}
			info, _ := e.Info()
			var size int64
			if info != nil {
				size = info.Size()
			}
			roms = append(roms, ROM{
				Name:   name,
				Path:   filepath.Join(sysPath, e.Name()),
				Size:   size,
				System: s.Name(),
			})
		}
	}
	return roms
}

// handleSearch returns every ROM whose name contains ?q= across all systems.
func handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	q := r.URL.Query().Get("q")
	if len(strings.TrimSpace(q)) < 2 {
		jsonError(w, "Query too short (min 2 chars)", http.StatusBadRequest)
		return
	}

	roms := collectRoms("", q)
	sort.Slice(roms, func(i, j int) bool { return roms[i].Name < roms[j].Name })
	if len(roms) > 300 {
		roms = roms[:300]
	}
	jsonOK(w, roms)
}

// handleRandom picks one random ROM (optionally within ?system=) — "Surprise me".
func handleRandom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	roms := collectRoms(r.URL.Query().Get("system"), "")
	if len(roms) == 0 {
		jsonError(w, "No ROMs found", http.StatusNotFound)
		return
	}
	jsonOK(w, roms[rand.Intn(len(roms))])
}
