package main

import (
	"encoding/binary"
	"encoding/json"
	"net/http"
	"os"
	"time"
)

// inputEvent matches the 16-byte Linux input_event struct on 32-bit ARM.
// tv_sec (4) + tv_usec (4) + type (2) + code (2) + value (4)
type inputEvent struct {
	TimeSec  int32
	TimeUsec int32
	Type     uint16
	Code     uint16
	Value    int32
}

const (
	evKey = 1
	evSyn = 0

	// Key codes from linux/input.h (matches Onion keymap_hw.h)
	keyUp       = 103
	keyDown     = 108
	keyLeft     = 105
	keyRight    = 106
	keyA        = 57  // KEY_SPACE
	keyB        = 29  // KEY_LEFTCTRL
	keyX        = 42  // KEY_LEFTSHIFT
	keyY        = 56  // KEY_LEFTALT
	keyL1       = 18  // KEY_E
	keyR1       = 20  // KEY_T
	keyL2       = 15  // KEY_TAB
	keyR2       = 14  // KEY_BACKSPACE
	keySelect   = 97  // KEY_RIGHTCTRL
	keyStart    = 28  // KEY_ENTER
	keyMenu     = 1   // KEY_ESC
	keyVolumeUp = 115 // KEY_VOLUMEUP
	keyVolumeDn = 114 // KEY_VOLUMEDOWN
)

var buttonMap = map[string]uint16{
	"up":        keyUp,
	"down":      keyDown,
	"left":      keyLeft,
	"right":     keyRight,
	"a":         keyA,
	"b":         keyB,
	"x":         keyX,
	"y":         keyY,
	"l1":        keyL1,
	"r1":        keyR1,
	"l2":        keyL2,
	"r2":        keyR2,
	"select":    keySelect,
	"start":     keyStart,
	"menu":      keyMenu,
	"volume_up": keyVolumeUp,
	"volume_dn": keyVolumeDn,
}

func handleInputPress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Button string `json:"button"`
		Action string `json:"action"` // "press", "release", or "tap" (default)
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	code, ok := buttonMap[req.Button]
	if !ok {
		jsonError(w, "Unknown button: "+req.Button, http.StatusBadRequest)
		return
	}

	action := req.Action
	if action == "" {
		action = "tap"
	}

	if err := sendInputEvent(code, action); err != nil {
		jsonError(w, "Input error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]string{"ok": "1"})
}

// macroStep is one entry in a button sequence.
type macroStep struct {
	Button string `json:"button"`
	Action string `json:"action"`   // press/release/tap (default: tap)
	Delay  int    `json:"delay_ms"` // pause AFTER this step
}

// handleInputMacro plays a sequence of button presses server-side — e.g. combos
// or the Konami code. POST /api/input/macro {"steps":[{button,action,delay_ms}]}
func handleInputMacro(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Steps []macroStep `json:"steps"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if len(req.Steps) == 0 || len(req.Steps) > 64 {
		jsonError(w, "Macro must have 1–64 steps", http.StatusBadRequest)
		return
	}
	for _, s := range req.Steps {
		if _, ok := buttonMap[s.Button]; !ok {
			jsonError(w, "Unknown button: "+s.Button, http.StatusBadRequest)
			return
		}
	}

	// Play asynchronously so a long macro doesn't hold the HTTP connection.
	go func(steps []macroStep) {
		for _, s := range steps {
			action := s.Action
			if action == "" {
				action = "tap"
			}
			sendInputEvent(buttonMap[s.Button], action)
			d := s.Delay
			if d < 0 {
				d = 0
			}
			if d > 2000 {
				d = 2000
			}
			if d == 0 {
				d = 60 // sensible default gap between buttons
			}
			time.Sleep(time.Duration(d) * time.Millisecond)
		}
	}(req.Steps)

	jsonOK(w, map[string]interface{}{"playing": len(req.Steps)})
}

func sendInputEvent(code uint16, action string) error {
	f, err := os.OpenFile("/dev/input/event0", os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	now := time.Now()
	sec := int32(now.Unix())
	usec := int32(now.Nanosecond() / 1000)

	writeEv := func(evType uint16, evCode uint16, val int32) error {
		ev := inputEvent{TimeSec: sec, TimeUsec: usec, Type: evType, Code: evCode, Value: val}
		return binary.Write(f, binary.LittleEndian, &ev)
	}
	syn := func() error { return writeEv(evSyn, 0, 0) }

	switch action {
	case "press":
		if err := writeEv(evKey, code, 1); err != nil {
			return err
		}
		return syn()
	case "release":
		if err := writeEv(evKey, code, 0); err != nil {
			return err
		}
		return syn()
	default: // "tap": press + release
		if err := writeEv(evKey, code, 1); err != nil {
			return err
		}
		if err := syn(); err != nil {
			return err
		}
		if err := writeEv(evKey, code, 0); err != nil {
			return err
		}
		return syn()
	}
}
