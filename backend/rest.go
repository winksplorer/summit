// summit backend/basicApi.go - simple http endpoint shit. more complex actions should be in comm.

package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// inits http handlers
func REST_Init() {
	log.Println("REST_Init: Init REST handlers.")
	http.HandleFunc("/", REST_Origin)
	http.HandleFunc("/api/login", REST_Login)
	http.HandleFunc("/api/logout", REST_Logout)
	http.HandleFunc("/api/hostname", REST_Hostname)
	http.HandleFunc("/api/authenticated", REST_Authenticated)
	http.HandleFunc("/api/suid", REST_SUID)
	http.HandleFunc("/api/pty", REST_Pty)
	http.HandleFunc("/api/comm", REST_Comm)
}

// file serving and templates. handles /.
func REST_Origin(w http.ResponseWriter, r *http.Request) {
	// only get filename
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		path = "index.html"
	}

	// if it doesn't need templating, then directly serve it
	if !strings.HasSuffix(path, ".html") || path == "index.html" || path == "admin.html" {
		http.FileServer(http.Dir(frontendDir)).ServeHTTP(w, r)
		return
	}

	pageName := strings.TrimSuffix(path, ".html")

	// template together base + the page
	tmpl, err := template.ParseFiles(fmt.Sprintf("%s/template/base.html", frontendDir), fmt.Sprintf("%s/template/%s.html", frontendDir, pageName))
	if err != nil {
		H_ISE(w, fmt.Sprintf("template parse error for %s", path), err)
		return
	}

	// get user. if not found then redirect to login
	sc, err := r.Cookie("s")
	if err != nil {
		http.Redirect(w, r, "/?err=inv", http.StatusFound)
		return
	}

	authsMu.RLock()
	defer authsMu.RUnlock()

	u, ok := auths[sc.Value]
	if !ok {
		http.Redirect(w, r, "/?err=inv", http.StatusFound)
		return
	}

	// create json
	data, err := json.Marshal(u.config)
	if err != nil {
		H_ISE(w, "couldn't represent config as json", err)
		return
	}

	err = tmpl.ExecuteTemplate(w, pageName, map[string]interface{}{
		"Title":    pageName + " - " + hostname,
		"Config":   template.JS(data),
		"Hostname": hostname,

		// page specific shit
		"BuildString": buildString,
	})
	if err != nil {
		H_ISE(w, fmt.Sprintf("template exec error for %s", path), err)
		return
	}
}

// http wrapper for A_Authenticated(w, r). handles /api/authenticated.
func REST_Authenticated(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if A_Authenticated(w, r) {
		http.Error(w, "OK", http.StatusOK)
		return
	}

	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

// handles /api/hostname
func REST_Hostname(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintln(w, hostname)
}

// handles /api/suid
func REST_SUID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// basic security checks
	if !A_Authenticated(w, r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if os.Geteuid() != 0 {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		H_ISE(w, "couldn't read /api/sudo data", err)
		return
	}
	defer r.Body.Close()

	// parse
	var decoded map[string]string
	if err := json.Unmarshal(body, &decoded); err != nil {
		log.Println("error: failed to parse /api/sudo data:", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// authenticate as root with PAM
	if err := H_PamAuth("passwd", "root", decoded["password"]); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// see what command we need
	cmdStr, ok := allowedSudoCommands[decoded["operation"]]
	if !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// execute command
	cmd := exec.Command(cmdStr)
	err = cmd.Run()
	if err != nil {
		H_ISE(w, "couldn't run command as root", err)
		return
	}

	// we good
	http.Error(w, "OK", http.StatusOK)
}
