package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
)

// Miyoo Mini framebuffer: 640x480 BGR565 (16-bit per pixel)
// Physical framebuffer is triple-buffered (640×1440 virtual, yoffset ∈ {0,480,960})
const (
	fbWidth  = 640
	fbHeight = 480
	fbBPP    = 2 // bytes per pixel (BGR565)
	fbSize   = fbWidth * fbHeight * fbBPP
)

func handleScreenshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	img, err := captureFramebuffer()
	if err != nil {
		jsonError(w, "Screenshot failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-cache, no-store")
	png.Encode(w, img)
}

func captureFramebuffer() (*image.RGBA, error) {
	f, err := os.Open(FbDevice)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Seek to the currently displayed buffer (FBIOGET_VSCREENINFO ioctl).
	// Without this we sometimes read a buffer mid-render → "old TV" tearing effect.
	if off := getFbYOffset(f); off > 0 {
		f.Seek(off, 0)
	}

	buf := make([]byte, fbSize)
	n, err := io.ReadFull(f, buf)
	if err != nil && n == 0 {
		return nil, err
	}

	w, h := fbWidth, fbHeight
	pixels := n / fbBPP
	if pixels < fbWidth*fbHeight && pixels >= 320*240 {
		w, h = 320, 240
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := (y*w + x) * fbBPP
			if idx+1 >= n {
				break
			}
			pixel := binary.LittleEndian.Uint16(buf[idx : idx+2])
			r8, g8, b8 := rgb565ToRGB888(pixel)
			// Miyoo Mini framebuffer is rotated 180° relative to display orientation.
			// Flip both X and Y axes to produce an upright image.
			img.SetRGBA(w-1-x, h-1-y, color.RGBA{R: r8, G: g8, B: b8, A: 255})
		}
	}
	return img, nil
}

// rgb565ToRGB888 converts a BGR565 pixel (Miyoo Mini format) to RGB888.
// The Miyoo Mini stores blue in the high bits and red in the low bits.
func rgb565ToRGB888(pixel uint16) (uint8, uint8, uint8) {
	b5 := (pixel >> 11) & 0x1F
	g6 := (pixel >> 5) & 0x3F
	r5 := pixel & 0x1F
	return uint8((r5 * 255) / 31), uint8((g6 * 255) / 63), uint8((b5 * 255) / 31)
}

// screenshotBase64 returns the framebuffer as a JPEG data URI.
// JPEG is ~5x faster to encode than PNG on ARM — gives smoother live preview.
func screenshotBase64() string {
	img, err := captureFramebuffer()
	if err != nil {
		return ""
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		return ""
	}
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}
