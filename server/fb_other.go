//go:build !linux

package main

import "os"

type fbInfo struct {
	yoffset  int64
	bpp      int
	redOff   uint32
	redLen   uint32
	greenOff uint32
	greenLen uint32
	blueOff  uint32
	blueLen  uint32
}

// getFbInfo stub for non-Linux builds (local dev on Windows/macOS).
func getFbInfo(_ *os.File) fbInfo {
	return fbInfo{bpp: 16, redOff: 11, redLen: 5, greenOff: 5, greenLen: 6, blueOff: 0, blueLen: 5}
}
