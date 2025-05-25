package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

var allowedSudoCommands = map[string]string{
	"reboot":   "/sbin/reboot",
	"poweroff": "/sbin/poweroff",
}

// struct to sudo request data
type SudoRequest struct {
	Password  string `json:"password"`
	Operation string `json:"operation"`
}

// handles /api/get-hostname
func getHostnameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		hostname, err := os.Hostname()
		if err != nil {
			log.Println("couldn't get hostname:", err)
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
			log.Println("couldn't get cpu info:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		virtualMem, err := mem.VirtualMemory()
		if err != nil {
			log.Println("couldn't get memory info:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		var usageValue float64
		var usageUnit string
		const (
			kb = 1024
			mb = kb * 1024
			gb = mb * 1024
		)

		switch {
		case virtualMem.Used >= gb:
			usageValue = float64(virtualMem.Used) / float64(gb)
			usageUnit = "g"
		case virtualMem.Used >= mb:
			usageValue = float64(virtualMem.Used) / float64(mb)
			usageUnit = "m"
		case virtualMem.Used >= kb:
			usageValue = float64(virtualMem.Used) / float64(kb)
			usageUnit = "k"
		default:
			usageValue = float64(virtualMem.Used)
			usageUnit = ""
		}

		stats := map[string]interface{}{
			"memoryTotal":     humanReadable(virtualMem.Total),
			"memoryUsage":     fmt.Sprintf("%.1f", usageValue),
			"memoryUsageUnit": usageUnit,
			"processorUsage":  percentages[0],
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}
}

// handles /api/server-pages
func serverPagesHandler(w http.ResponseWriter, r *http.Request) {
	if !authenticated(w, r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method == http.MethodGet {
		dict := [][2]string{
			{"term.html", "terminal"},
			{"log.html", "logging"},
			{"stor.html", "storage"},
			{"net.html", "networking"},
			{"container.html", "containers"},
			{"services.html", "services"},
			{"updates.html", "updates"},
			{"config.html", "settings"},
		}

		jsonData, err := json.Marshal(dict)
		if err != nil {
			log.Println("couldn't format json:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(jsonData))
	}
}

// handles /api/buildstring
func buildstringHandler(w http.ResponseWriter, r *http.Request) {
	if !authenticated(w, r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method == http.MethodGet {
		fmt.Fprintf(w, "summit v%s (built on %s)", Version, BuildDate)
	}
}

// handles /api/sudo
func sudoHandler(w http.ResponseWriter, r *http.Request) {
	if !authenticated(w, r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method == http.MethodPost {
		if os.Geteuid() != 0 {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Parse the JSON request body
		var req SudoRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Println("error: failed to parse /api/sudo JSON:", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// authenticate as root with PAM
		if err := pamAuth("passwd", "root", req.Password); err != nil {
			http.Redirect(w, r, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// see what command we need
		cmdStr, ok := allowedSudoCommands[req.Operation]
		if !ok {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// execute command
		cmd := exec.Command(cmdStr)
		err := cmd.Run()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}

		// we good
		fmt.Fprint(w, "OK")
	}
}
