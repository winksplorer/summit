package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
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
	if !authenticated(w, r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method == http.MethodGet {
		virtualMem, err := mem.VirtualMemory()
		if err != nil {
			fmt.Println("couldn't get memory info:", err)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "mem %s/%s", humanReadable(virtualMem.Used), humanReadable(virtualMem.Total))
	}
}

// handles /api/stat-cpu
func statCpuHandler(w http.ResponseWriter, r *http.Request) {
	if !authenticated(w, r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method == http.MethodGet {
		percentages, err := cpu.Percent(0, false)
		if err != nil {
			fmt.Println("couldn't get cpu info:", err)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "cpu %d%%", int(percentages[0]))
	}
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

// generate random b64 str
func randomBase64String(length int) (string, error) {
	numBytes := (length * 3) / 4
	randomBytes := make([]byte, numBytes)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(randomBytes)[:length], nil
}
