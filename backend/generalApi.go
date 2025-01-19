package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"syscall"

	"github.com/shirou/gopsutil/v3/cpu"
)

// handles /api/get-hostname
func getHostnameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		hostname, err := os.Hostname()
		if err != nil {
			fmt.Println("couldn't get hostname:", err)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, hostname)
	}
}

// handles /api/stat-memory
func statMemoryHandler(w http.ResponseWriter, r *http.Request) {
	if !authenticated(strings.Split(r.RemoteAddr, ":")[0]) {
		return
	}

	if r.Method == http.MethodGet {
		var sysInfo syscall.Sysinfo_t

		// Get system memory info
		if err := syscall.Sysinfo(&sysInfo); err != nil {
			fmt.Println("couldn't get memory info:", err)
		}

		// Total and free memory in bytes
		totalMemory := sysInfo.Totalram
		freeMemory := sysInfo.Freeram

		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "mem %s/%s", humanReadable(totalMemory-freeMemory), humanReadable(totalMemory))
	}
}

// handles /api/stat-cpu
func statCpuHandler(w http.ResponseWriter, r *http.Request) {
	if !authenticated(strings.Split(r.RemoteAddr, ":")[0]) {
		return
	}

	if r.Method == http.MethodGet {
		percentages, err := cpu.Percent(0, false)
		if err != nil {
			fmt.Println("couldn't get cpu info:", err)
		}

		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "cpu %d%%", int(percentages[0]))
	}
}

// removes item from string slice
func remove(slice []string, value string) []string {
	var result []string
	for _, item := range slice {
		if item != value {
			result = append(result, item)
		}
	}
	return result
}

// human readable byte sizes
func humanReadable(bytes uint64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)

	switch {
	case bytes >= gb:
		return fmt.Sprintf("%.1fg", float64(bytes)/float64(gb))
	case bytes >= mb:
		return fmt.Sprintf("%.1fm", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.1fk", float64(bytes)/float64(kb))
	default:
		return fmt.Sprintf("%d", bytes)
	}
}
