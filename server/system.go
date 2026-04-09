package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"math"
	"time"
)

// cpuUsageCache stores the last computed CPU % as a uint64 (IEEE 754 bits).
// Updated by a background goroutine so readCPUUsage() is non-blocking.
var cpuUsageCache uint64

func init() {
	go func() {
		for {
			s1 := readCPUStat()
			time.Sleep(200 * time.Millisecond)
			s2 := readCPUStat()
			var pct float64
			if s1 != nil && s2 != nil {
				idle := float64(s2[3] - s1[3])
				total := float64(sum(s2) - sum(s1))
				if total > 0 {
					pct = (1.0 - idle/total) * 100.0
				}
			}
			atomic.StoreUint64(&cpuUsageCache, math.Float64bits(pct))
			time.Sleep(1800 * time.Millisecond)
		}
	}()
}

type SystemInfo struct {
	CPU     float64 `json:"cpu_percent"`
	RAM     RAMInfo `json:"ram"`
	Battery BatInfo `json:"battery"`
	IP      string  `json:"ip"`
	Uptime  string  `json:"uptime"`
	CPUFreq int     `json:"cpu_freq_mhz"`
}

type RAMInfo struct {
	Total     int `json:"total_mb"`
	Used      int `json:"used_mb"`
	Available int `json:"available_mb"`
}

type BatInfo struct {
	Percent  int    `json:"percent"`
	Charging bool   `json:"charging"`
	Voltage  string `json:"voltage"`
}

func handleSystem(w http.ResponseWriter, r *http.Request) {
	info := SystemInfo{
		CPU:     readCPUUsage(),
		RAM:     readRAM(),
		Battery: readBattery(),
		IP:      getLocalIP(),
		Uptime:  readUptime(),
		CPUFreq: readCPUFreq(),
	}
	jsonOK(w, info)
}

func readCPUUsage() float64 {
	return math.Float64frombits(atomic.LoadUint64(&cpuUsageCache))
}

func readCPUStat() []int64 {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)[1:]
			vals := make([]int64, len(fields))
			for i, f := range fields {
				vals[i], _ = strconv.ParseInt(f, 10, 64)
			}
			return vals
		}
	}
	return nil
}

func sum(vals []int64) int64 {
	var s int64
	for _, v := range vals {
		s += v
	}
	return s
}

func readRAM() RAMInfo {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return RAMInfo{}
	}
	defer f.Close()

	mem := map[string]int{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			key := strings.TrimSuffix(parts[0], ":")
			val, _ := strconv.Atoi(parts[1])
			mem[key] = val
		}
	}

	totalKB := mem["MemTotal"]
	availKB := mem["MemAvailable"]
	if availKB == 0 {
		availKB = mem["MemFree"] + mem["Buffers"] + mem["Cached"]
	}
	usedKB := totalKB - availKB

	return RAMInfo{
		Total:     totalKB / 1024,
		Used:      usedKB / 1024,
		Available: availKB / 1024,
	}
}

func readBattery() BatInfo {
	// Try Miyoo Mini Plus AXP (via /sys/class/power_supply)
	paths := []string{
		"/sys/class/power_supply/axp2202-battery",
		"/sys/class/power_supply/battery",
		"/sys/class/power_supply/BAT0",
	}

	for _, base := range paths {
		capBytes, err := os.ReadFile(base + "/capacity")
		if err != nil {
			continue
		}
		cap, _ := strconv.Atoi(strings.TrimSpace(string(capBytes)))

		statusBytes, _ := os.ReadFile(base + "/status")
		status := strings.TrimSpace(string(statusBytes))

		voltBytes, _ := os.ReadFile(base + "/voltage_now")
		volt := strings.TrimSpace(string(voltBytes))
		if v, err := strconv.ParseFloat(volt, 64); err == nil {
			volt = fmt.Sprintf("%.2fV", v/1000000.0)
		}

		return BatInfo{
			Percent:  cap,
			Charging: strings.EqualFold(status, "Charging"),
			Voltage:  volt,
		}
	}

	// Fallback: try reading from Miyoo AXP GPIO
	gpioVal, err := os.ReadFile("/sys/devices/gpiochip0/gpio/gpio59/value")
	if err == nil {
		charging := strings.TrimSpace(string(gpioVal)) == "1"
		return BatInfo{Percent: -1, Charging: charging, Voltage: "N/A"}
	}

	return BatInfo{Percent: -1, Charging: false, Voltage: "N/A"}
}

func readUptime() string {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return "unknown"
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return "unknown"
	}
	secs, _ := strconv.ParseFloat(fields[0], 64)
	d := time.Duration(secs) * time.Second
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", h, m)
}

func readCPUFreq() int {
	data, err := os.ReadFile("/sys/devices/system/cpu/cpu0/cpufreq/scaling_cur_freq")
	if err != nil {
		return 0
	}
	khz, _ := strconv.Atoi(strings.TrimSpace(string(data)))
	return khz / 1000
}
