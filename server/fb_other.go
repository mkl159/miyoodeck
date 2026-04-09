//go:build !linux

package main

import "os"

// getFbYOffset is a no-op stub for non-Linux builds (local dev / CI).
func getFbYOffset(_ *os.File) int64 { return 0 }
