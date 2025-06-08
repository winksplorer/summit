package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
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
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Println("couldn't get hostname:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintln(w, hostname)
}

// handles /api/sudo
func sudoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if !authenticated(w, r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

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
	http.Error(w, "OK", http.StatusOK)
}
