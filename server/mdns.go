package main

// Minimal mDNS responder — répond aux requêtes DNS pour "miyoodeck.local"
// Implémenté sans dépendance externe, multicast UDP 224.0.0.251:5353

import (
	"encoding/binary"
	"log"
	"net"
	"strings"
)

const mdnsAddr = "224.0.0.251:5353"

// startMDNS annonce miyoodeck.local → ip sur le réseau local.
func startMDNS(ip string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("mDNS panic (non-fatal): %v", r)
		}
	}()

	parsed := net.ParseIP(ip)
	if parsed == nil {
		log.Printf("mDNS: invalid IP %q, skipping", ip)
		return
	}
	parsedIP := parsed.To4()
	if parsedIP == nil {
		log.Printf("mDNS: not an IPv4 address: %q, skipping", ip)
		return
	}

	addr, err := net.ResolveUDPAddr("udp4", mdnsAddr)
	if err != nil {
		log.Printf("mDNS: resolve error: %v", err)
		return
	}
	conn, err := net.ListenMulticastUDP("udp4", nil, addr)
	if err != nil {
		// mDNS non disponible — pas critique
		log.Printf("mDNS: listen error (non-fatal): %v", err)
		return
	}
	defer conn.Close()

	log.Printf("mDNS: listening, announcing miyoodeck.local → %s", parsedIP)

	buf := make([]byte, 1500)
	for {
		n, src, err := conn.ReadFromUDP(buf)
		if err != nil {
			return
		}
		if isQueryForMiyooDeck(buf[:n]) {
			resp := buildMDNSResponse(parsedIP)
			conn.WriteToUDP(resp, src)
			// Also send to multicast so all devices on LAN see it
			conn.WriteToUDP(resp, addr)
		}
	}
}

// isQueryForMiyooDeck checks if the DNS packet contains a question for "miyoodeck.local".
func isQueryForMiyooDeck(pkt []byte) bool {
	if len(pkt) < 12 {
		return false
	}
	// QR bit must be 0 (query), OPCODE=0 (standard)
	flags := binary.BigEndian.Uint16(pkt[2:4])
	if flags&0x8000 != 0 {
		return false // it's a response
	}
	// Search for "miyoodeck" in the label sequence
	lower := strings.ToLower(string(pkt))
	return strings.Contains(lower, "miyoodeck")
}

// buildMDNSResponse builds a minimal DNS A-record response for miyoodeck.local.
func buildMDNSResponse(ip net.IP) []byte {
	//  DNS header: ID=0, QR=1 (response), AA=1 (authoritative), no questions
	hdr := []byte{
		0x00, 0x00, // ID
		0x84, 0x00, // Flags: QR=1, AA=1, no error
		0x00, 0x00, // QDCOUNT = 0
		0x00, 0x01, // ANCOUNT = 1
		0x00, 0x00, // NSCOUNT = 0
		0x00, 0x00, // ARCOUNT = 0
	}

	// Name: miyoodeck.local (encoded as DNS labels)
	name := encodeDNSName("miyoodeck.local")

	// Answer RR: TYPE=A(1), CLASS=IN(1) | FLUSH(0x8000), TTL=120, RDATA=4 bytes IP
	rr := []byte{}
	rr = append(rr, name...)
	rr = append(rr, 0x00, 0x01) // TYPE A
	rr = append(rr, 0x80, 0x01) // CLASS IN + cache-flush bit
	rr = append(rr, 0x00, 0x00, 0x00, 0x78) // TTL = 120s
	rr = append(rr, 0x00, 0x04) // RDLENGTH = 4
	rr = append(rr, ip[0], ip[1], ip[2], ip[3]) // RDATA

	return append(hdr, rr...)
}

// encodeDNSName encodes a dotted name (e.g. "miyoodeck.local") as DNS labels.
func encodeDNSName(name string) []byte {
	var out []byte
	for _, label := range strings.Split(strings.TrimSuffix(name, "."), ".") {
		out = append(out, byte(len(label)))
		out = append(out, []byte(label)...)
	}
	out = append(out, 0x00) // root label
	return out
}
