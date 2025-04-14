package main

import (
	"encoding/json"
	"fmt"
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

// handles /api/server-pages
func serverPagesHandler(w http.ResponseWriter, r *http.Request) {
	if !authenticated(w, r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method == http.MethodGet {
		var dict [][2]string

		if Edition == "RH" {
			dict = [][2]string{
				{"term.html", "terminal"},
				{"log.html", "logging"},
				{"stor.html", "storage"},
				{"net.html", "networking"},
				{"vm.html", "virtual machines"},
				{"services.html", "services"},
				{"updates.html", "updates"},
				{"config.html", "settings"},
			}
		} else {
			dict = [][2]string{
				{"term.html", "terminal"},
				{"log.html", "logging"},
				{"stor.html", "storage"},
				{"net.html", "networking"},
				{"services.html", "services"},
				{"updates.html", "updates"},
				{"config.html", "settings"},
			}
		}

		jsonData, err := json.Marshal(dict)
		if err != nil {
			fmt.Println("couldn't format json:", err)
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
		var editionName = "guests"
		if Edition == "RH" {
			editionName = "hosts"
		}

		fmt.Fprintf(w, "summit for %s v%s (built on %s)", editionName, Version, BuildDate)
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
			fmt.Println("error: failed to parse /api/sudo JSON:", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// authenticate as root with PAM
		if err := PAMAuth("passwd", "root", req.Password); err != nil {
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
