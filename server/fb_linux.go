//go:build linux

package main

import (
	"os"
	"syscall"
	"unsafe"
)

const fbIOGetVScreenInfo = 0x4600

// fbInfo holds the framebuffer parameters we need to correctly capture a frame.
type fbInfo struct {
	yoffset  int64 // byte offset to the currently displayed buffer
	bpp      int   // bits per pixel (16 or 32)
	redOff   uint32
	redLen   uint32
	greenOff uint32
	greenLen uint32
	blueOff  uint32
	blueLen  uint32
}

// getFbInfo queries the kernel for the real framebuffer pixel format and
// current display buffer offset.  Works for both 16-bit RGB565 (original MM)
// and 32-bit ABGR8888 (Miyoo Mini Plus), regardless of BGR vs RGB channel order.
// Returns safe defaults on error (16-bit RGB565 from offset 0).
func getFbInfo(f *os.File) fbInfo {
	// fb_var_screeninfo is 160 bytes on 32-bit ARM Linux.
	// Field offsets (all uint32, little-endian):
	//   0  xres
	//   4  yres
	//   8  xres_virtual
	//  12  yres_virtual
	//  16  xoffset
	//  20  yoffset
	//  24  bits_per_pixel
	//  28  grayscale
	//  32  red.offset   36 red.length   40 red.msb_right
	//  44  green.offset 48 green.length 52 green.msb_right
	//  56  blue.offset  60 blue.length  64 blue.msb_right
	var buf [160]byte
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		f.Fd(),
		fbIOGetVScreenInfo,
		uintptr(unsafe.Pointer(&buf[0])),
	)
	if errno != 0 {
		return defaultFbInfo()
	}

	u32 := func(off int) uint32 {
		return uint32(buf[off]) | uint32(buf[off+1])<<8 |
			uint32(buf[off+2])<<16 | uint32(buf[off+3])<<24
	}

	xres := u32(0)
	yoff := u32(20)
	bpp := u32(24)

	if bpp == 0 {
		return defaultFbInfo()
	}

	return fbInfo{
		yoffset:  int64(yoff) * int64(xres) * int64(bpp/8),
		bpp:      int(bpp),
		redOff:   u32(32),
		redLen:   u32(36),
		greenOff: u32(44),
		greenLen: u32(48),
		blueOff:  u32(56),
		blueLen:  u32(60),
	}
}

func defaultFbInfo() fbInfo {
	// RGB565 fallback (Miyoo Mini original)
	return fbInfo{bpp: 16, redOff: 11, redLen: 5, greenOff: 5, greenLen: 6, blueOff: 0, blueLen: 5}
}
