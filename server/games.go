package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type System struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	RomCount int    `json:"rom_count"`
	Icon     string `json:"icon,omitempty"`
}

type ROM struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	System string `json:"system"`
}

// Extensions that are valid ROM files
var romExtensions = map[string]bool{
	".gba": true, ".gbc": true, ".gb": true,
	".sfc": true, ".smc": true, ".nes": true,
	".md": true, ".gen": true, ".sms": true,
	".pce": true, ".ngp": true, ".ws": true,
	".ggg": true, ".gg": true, ".32x": true,
	".iso": true, ".pbp": true, ".cue": true,
	".bin": true, ".img": true, ".nds": true,
	".n64": true, ".z64": true, ".v64": true,
	".zip": true, ".7z": true,
}

func handleSystems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	entries, err := os.ReadDir(RomsDir)
	if err != nil {
		jsonError(w, "Cannot read Roms directory: "+err.Error(), http.StatusNotFound)
		return
	}

	var systems []System
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}

		sysPath := filepath.Join(RomsDir, e.Name())
		count := countRoms(sysPath)

		sys := System{
			Name:     e.Name(),
			Path:     sysPath,
			RomCount: count,
		}

		// Try to find icon
		iconPath := fmt.Sprintf("/mnt/SDCARD/Icons/Default/emu/%s.png",
			strings.ToLower(e.Name()))
		if _, err := os.Stat(iconPath); err == nil {
			sys.Icon = iconPath
		}

		systems = append(systems, sys)
	}

	sort.Slice(systems, func(i, j int) bool {
		return systems[i].Name < systems[j].Name
	})

	jsonOK(w, systems)
}

func countRoms(dir string) int {
	count := 0
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	for _, e := range entries {
		if !e.IsDir() {
			ext := strings.ToLower(filepath.Ext(e.Name()))
			if romExtensions[ext] {
				count++
			}
		}
	}
	return count
}

func handleRoms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	system := r.URL.Query().Get("system")
	if system == "" {
		jsonError(w, "Missing system parameter", http.StatusBadRequest)
		return
	}

	// Prevent path traversal
	system = filepath.Base(system)
	sysPath := filepath.Join(RomsDir, system)

	if _, err := os.Stat(sysPath); os.IsNotExist(err) {
		jsonError(w, "System not found", http.StatusNotFound)
		return
	}

	entries, err := os.ReadDir(sysPath)
	if err != nil {
		jsonError(w, "Cannot read system directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var roms []ROM
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if !romExtensions[ext] {
			continue
		}
		info, _ := e.Info()
		var size int64
		if info != nil {
			size = info.Size()
		}
		roms = append(roms, ROM{
			Name:   strings.TrimSuffix(e.Name(), filepath.Ext(e.Name())),
			Path:   filepath.Join(sysPath, e.Name()),
			Size:   size,
			System: system,
		})
	}

	sort.Slice(roms, func(i, j int) bool {
		return roms[i].Name < roms[j].Name
	})

	jsonOK(w, roms)
}

func handleLaunch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RomPath string `json:"rom_path"`
		System  string `json:"system"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.RomPath == "" {
		jsonError(w, "Missing rom_path", http.StatusBadRequest)
		return
	}

	req.RomPath = filepath.Clean(req.RomPath)
	if !withinSD(req.RomPath) {
		jsonError(w, "Access denied", http.StatusForbidden)
		return
	}

	if _, err := os.Stat(req.RomPath); os.IsNotExist(err) {
		jsonError(w, "ROM file not found", http.StatusNotFound)
		return
	}

	// Find the emulator launch script for this system
	launchCmd, err := buildLaunchCommand(req.RomPath, req.System)
	if err != nil {
		jsonError(w, "Cannot find emulator for system: "+err.Error(), http.StatusNotFound)
		return
	}

	// Fix #5: if a game (retroarch) is already running, send a safe quit signal
	// instead of hard-killing — avoids save corruption
	if isGameRunning() {
		// Ask RetroArch to quit cleanly via its network command interface
		exec.Command("sh", "-c",
			`echo -e "QUIT\n" | nc -u -w1 127.0.0.1 55355 2>/dev/null || `+
				`killall -15 retroarch 2>/dev/null`).Run()
		// Give it 2 seconds to save and exit
		time.Sleep(2 * time.Second)
	}

	// Write the launch command
	cmdFile := SysDir + "/cmd_to_run.sh"
	if err := os.WriteFile(cmdFile, []byte(launchCmd), 0755); err != nil {
		jsonError(w, "Cannot write launch command: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Kill MainUI to trigger game launch (Onion will pick up cmd_to_run.sh)
	exec.Command("killall", "-9", "MainUI").Run()

	jsonOK(w, map[string]string{
		"message": "Launching: " + filepath.Base(req.RomPath),
		"command": launchCmd,
	})
}

func buildLaunchCommand(romPath, system string) (string, error) {
	if system == "" {
		system = filepath.Base(filepath.Dir(romPath))
	}

	// Look for emulator config in /mnt/SDCARD/Emu/
	emuDir := filepath.Join(EmuDir, system)
	configPath := filepath.Join(emuDir, "config.json")

	type EmuConfig struct {
		Launch string `json:"launch"`
	}

	var cfg EmuConfig
	data, err := os.ReadFile(configPath)
	if err == nil {
		json.Unmarshal(data, &cfg)
	}

	launchScript := filepath.Join(emuDir, "launch.sh")
	if _, err := os.Stat(launchScript); err == nil {
		return fmt.Sprintf("cd %q && sh launch.sh %q", emuDir, romPath), nil
	}

	// Fallback: try to use RetroArch directly
	// Find the core for this system
	corePath := findCore(system)
	if corePath != "" {
		return fmt.Sprintf(
			`cd /mnt/SDCARD/RetroArch && ./retroarch -v -L %q %q`,
			corePath, romPath,
		), nil
	}

	return "", fmt.Errorf("no emulator found for system %q", system)
}

// isGameRunning returns true if retroarch or a game emulator is currently active.
func isGameRunning() bool {
	out, err := exec.Command("pgrep", "-x", "retroarch").Output()
	return err == nil && len(out) > 0
}

func findCore(system string) string {
	// Map common system names to core names
	coreMap := map[string]string{
		"GBA":   "gpsp_libretro",
		"GBC":   "gambatte_libretro",
		"GB":    "gambatte_libretro",
		"SFC":   "snes9x2005_plus_libretro",
		"SNES":  "snes9x2005_plus_libretro",
		"FC":    "fceumm_libretro",
		"NES":   "fceumm_libretro",
		"MD":    "picodrive_libretro",
		"SMS":   "picodrive_libretro",
		"PS":    "pcsx_rearmed_libretro",
		"PSX":   "pcsx_rearmed_libretro",
		"PCE":   "mednafen_pce_fast_libretro",
		"NGP":   "mednafen_ngp_libretro",
		"GG":    "genesis_plus_gx_libretro",
		"WSWAN": "mednafen_wswan_libretro",
	}

	coreName, ok := coreMap[strings.ToUpper(system)]
	if !ok {
		return ""
	}

	corePath := fmt.Sprintf("/mnt/SDCARD/RetroArch/.retroarch/cores/%s.so", coreName)
	if _, err := os.Stat(corePath); err == nil {
		return corePath
	}
	return ""
}
