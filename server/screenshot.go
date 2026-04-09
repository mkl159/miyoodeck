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

// Miyoo Mini framebuffer dimensions — always 640×480 logical pixels.
// BPP (16 or 32) and channel layout are read at runtime from the kernel.
const (
	fbWidth  = 640
	fbHeight = 480
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

	// Read kernel pixel format + current display buffer offset (triple buffer).
	info := getFbInfo(f)
	bytesPerPixel := info.bpp / 8
	if bytesPerPixel < 1 {
		bytesPerPixel = 2
	}
	fbSize := fbWidth * fbHeight * bytesPerPixel

	// Seek to the currently displayed buffer to avoid "old TV" tearing.
	if info.yoffset > 0 {
		f.Seek(info.yoffset, 0)
	}

	buf := make([]byte, fbSize)
	n, err := io.ReadFull(f, buf)
	if err != nil && n == 0 {
		return nil, err
	}

	w, h := fbWidth, fbHeight
	pixels := n / bytesPerPixel
	if pixels < fbWidth*fbHeight && pixels >= 320*240 {
		w, h = 320, 240
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := (y*w + x) * bytesPerPixel
			if idx+bytesPerPixel > n {
				break
			}

			var pixel uint32
			switch bytesPerPixel {
			case 2:
				pixel = uint32(binary.LittleEndian.Uint16(buf[idx : idx+2]))
			case 4:
				pixel = binary.LittleEndian.Uint32(buf[idx : idx+4])
			}

			r8 := channelTo8bit(pixel, info.redOff, info.redLen)
			g8 := channelTo8bit(pixel, info.greenOff, info.greenLen)
			b8 := channelTo8bit(pixel, info.blueOff, info.blueLen)

			// Miyoo Mini framebuffer is rotated 180° relative to display orientation.
			// Flip both X and Y to produce an upright image.
			img.SetRGBA(w-1-x, h-1-y, color.RGBA{R: r8, G: g8, B: b8, A: 255})
		}
	}
	return img, nil
}

// channelTo8bit extracts a color channel of `length` bits at bit `offset`
// from a pixel value and scales it to 8 bits.
func channelTo8bit(pixel uint32, offset, length uint32) uint8 {
	if length == 0 {
		return 0
	}
	mask := uint32((1 << length) - 1)
	val := (pixel >> offset) & mask
	maxVal := mask
	return uint8((val * 255) / maxVal)
}

// screenshotBase64 returns the framebuffer as a JPEG data URI.
// JPEG is ~5× faster to encode than PNG on ARM — better live preview.
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
