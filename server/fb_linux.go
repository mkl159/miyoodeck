//go:build linux

package main

import (
	"os"
	"syscall"
	"unsafe"
)

// FBIOGET_VSCREENINFO reads the variable screen info from the framebuffer driver.
// Used to determine the current display buffer offset (triple buffering).
const fbIOGetVScreenInfo = 0x4600

// getFbYOffset returns the byte offset into /dev/fb0 where the currently
// displayed frame begins.  On the Miyoo Mini (triple buffer at yoffset ∈
// {0, 480, 960}), reading from this offset avoids the "old TV tearing" effect
// caused by reading a buffer that is simultaneously being rendered into.
// Returns 0 on any error (safe fallback: read from the start of the file).
func getFbYOffset(f *os.File) int64 {
	// fb_var_screeninfo is 160 bytes on 32-bit ARM Linux.
	// yoffset is a uint32 at byte offset 20.
	var info [160]byte
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		f.Fd(),
		fbIOGetVScreenInfo,
		uintptr(unsafe.Pointer(&info[0])),
	)
	if errno != 0 {
		return 0
	}
	yoff := uint32(info[20]) |
		uint32(info[21])<<8 |
		uint32(info[22])<<16 |
		uint32(info[23])<<24
	return int64(yoff) * int64(fbWidth) * int64(fbBPP)
}
