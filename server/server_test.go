package main

import (
	"encoding/binary"
	"testing"
)

func TestWithinSD(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"/mnt/SDCARD", true},
		{"/mnt/SDCARD/Roms/GBA/game.gba", true},
		{"/mnt/SDCARD/.tmp_update/config/x.cfg", true},
		{"/mnt/SDCARDevil", false}, // sibling must be rejected
		{"/mnt/SDCARD-backup", false},
		{"/etc/passwd", false},
		{"/", false},
		{"", false},
	}
	for _, c := range cases {
		if got := withinSD(c.path); got != c.want {
			t.Errorf("withinSD(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestHashPin(t *testing.T) {
	// Known SHA-256 vector for "1234".
	const want = "03ac674216f3e15c761ee1a5e255f067953623c8b388b4459e13f978d7c846f4"
	if got := hashPin("1234"); got != want {
		t.Errorf("hashPin(\"1234\") = %s, want %s", got, want)
	}
	if hashPin("1234") == hashPin("1235") {
		t.Error("different PINs produced the same hash")
	}
}

func TestChannelTo8bit(t *testing.T) {
	cases := []struct {
		name                  string
		pixel, offset, length uint32
		want                  uint8
	}{
		{"zero length", 0xffff, 0, 0, 0},
		{"5-bit full", 0x1f, 0, 5, 255},
		{"5-bit zero", 0x00, 0, 5, 0},
		{"6-bit full", 0x3f, 0, 6, 255},
		{"RGB565 red full", 0x1f << 11, 11, 5, 255},
		{"RGB565 green full", 0x3f << 5, 5, 6, 255},
	}
	for _, c := range cases {
		if got := channelTo8bit(c.pixel, c.offset, c.length); got != c.want {
			t.Errorf("%s: channelTo8bit(%#x,%d,%d) = %d, want %d",
				c.name, c.pixel, c.offset, c.length, got, c.want)
		}
	}
}

func TestParseVoltageFromAxp(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{`{"battery":85, "voltage":3950, "charging":0}`, "3.95V"},
		{`{"battery":85, "voltage":4100, "charging":3}`, "4.10V"},
		{`{"battery":85}`, "N/A"},
		{"garbage", "N/A"},
	}
	for _, c := range cases {
		if got := parseVoltageFromAxp(c.in); got != c.want {
			t.Errorf("parseVoltageFromAxp(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestEncodeDNSName(t *testing.T) {
	got := encodeDNSName("miyoodeck.local")
	want := []byte{9, 'm', 'i', 'y', 'o', 'o', 'd', 'e', 'c', 'k', 5, 'l', 'o', 'c', 'a', 'l', 0}
	if string(got) != string(want) {
		t.Errorf("encodeDNSName = %v, want %v", got, want)
	}
}

func TestIsQueryForMiyooDeck(t *testing.T) {
	// Build a minimal mDNS query containing "miyoodeck.local".
	query := make([]byte, 12) // header, all-zero flags => it's a query
	query = append(query, encodeDNSName("miyoodeck.local")...)
	query = append(query, 0x00, 0x01, 0x00, 0x01) // QTYPE=A, QCLASS=IN
	if !isQueryForMiyooDeck(query) {
		t.Error("expected query for miyoodeck.local to match")
	}

	// Same packet but with the QR (response) bit set must be ignored.
	resp := make([]byte, len(query))
	copy(resp, query)
	binary.BigEndian.PutUint16(resp[2:4], 0x8000)
	if isQueryForMiyooDeck(resp) {
		t.Error("response packet should not be treated as a query")
	}

	// Unrelated / too-short packets.
	if isQueryForMiyooDeck([]byte{1, 2, 3}) {
		t.Error("short packet should not match")
	}
	notUs := append(make([]byte, 12), encodeDNSName("printer.local")...)
	if isQueryForMiyooDeck(notUs) {
		t.Error("query for a different host should not match")
	}
}

func TestSum(t *testing.T) {
	if got := sum([]int64{1, 2, 3, 4}); got != 10 {
		t.Errorf("sum = %d, want 10", got)
	}
	if got := sum(nil); got != 0 {
		t.Errorf("sum(nil) = %d, want 0", got)
	}
}
