package main

import (
	"encoding/json"
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
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, hostname)
	}
}

// handles /api/stats
func statsHandler(w http.ResponseWriter, r *http.Request) {
	if !authenticated(w, r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method == http.MethodGet {
		percentages, err := cpu.Percent(0, false)
		if err != nil {
			fmt.Println("couldn't get cpu info:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		virtualMem, err := mem.VirtualMemory()
		if err != nil {
			fmt.Println("couldn't get memory info:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "<span>mem %s/%s</span> <span>cpu %d%%</span>", humanReadable(virtualMem.Used), humanReadable(virtualMem.Total), int(percentages[0]))
	}
}

// handles /api/server-euid
func servereuidHandler(w http.ResponseWriter, r *http.Request) {
	if !authenticated(w, r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method == http.MethodGet {
		euid := os.Geteuid()

		response := map[string]int{"euid": euid}
		jsonData, err := json.Marshal(response)
		if err != nil {
			fmt.Println("couldn't format json:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(jsonData))
	}
}

// handles /api/reboot
func rebootHandler(w http.ResponseWriter, r *http.Request) {
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

		virtualMem, err := mem.VirtualMemory()
		if err != nil {
			fmt.Println("couldn't get memory info:", err)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "<span>mem %s/%s</span> <span>cpu %d%%</span>", humanReadable(virtualMem.Used), humanReadable(virtualMem.Total), int(percentages[0]))
	}
}
