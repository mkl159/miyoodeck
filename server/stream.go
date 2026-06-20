package main

import (
	"fmt"
	"image/jpeg"
	"net/http"
	"time"
)

// handleStream serves the framebuffer as a real MJPEG stream
// (multipart/x-mixed-replace). Browsers render it natively in an <img>,
// giving a smooth, video-like live view without WebSocket base64 overhead.
//
//	GET /api/stream.mjpeg
//
// The stream ends as soon as the client disconnects (detected via the request
// context) so it never keeps reading the framebuffer for a closed tab.
func handleStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		jsonError(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	const boundary = "frame"
	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary="+boundary)
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Connection", "close")

	ctx := r.Context()
	// Slower cadence while a game runs to spare the CPU for the emulator.
	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		if isGameRunning() {
			ticker.Reset(300 * time.Millisecond)
		} else {
			ticker.Reset(150 * time.Millisecond)
		}

		img, err := captureFramebuffer()
		if err != nil {
			continue
		}

		if _, err := fmt.Fprintf(w, "--%s\r\nContent-Type: image/jpeg\r\n\r\n", boundary); err != nil {
			return
		}
		if err := jpeg.Encode(w, img, &jpeg.Options{Quality: 75}); err != nil {
			return
		}
		if _, err := fmt.Fprint(w, "\r\n"); err != nil {
			return
		}
		flusher.Flush()
	}
}
