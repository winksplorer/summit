// summit backend/rest.go - simple http endpoint shit. more complex actions should be in comm.

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

var (
	REST_Handlers = map[string]func(http.ResponseWriter, *http.Request){
		"/":                  REST_Origin,
		"/api/login":         REST_Login,
		"/api/logout":        REST_Logout,
		"/api/hostname":      REST_Hostname,
		"/api/authenticated": REST_Authenticated,
		"/api/suid":          REST_SUID,
		"/api/pty":           REST_Pty,
		"/api/comm":          REST_Comm,
	}

	REST_AllowedRootCommands = map[string]string{
		"reboot":   "/sbin/reboot",
		"poweroff": "/sbin/poweroff",
	}
)

// inits http handlers
func REST_Init() {
	log.Println("REST_Init: Init REST handlers.")
	for endpoint, handler := range REST_Handlers {
		http.HandleFunc(endpoint, handler)
	}
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
		http.FileServer(http.Dir(FrontendDir)).ServeHTTP(w, r)
		return
	}

	pageName := strings.TrimSuffix(path, ".html")

	// template together base + the page
	tmpl, err := template.ParseFiles(fmt.Sprintf("%s/template/base.html", FrontendDir), fmt.Sprintf("%s/template/%s.html", FrontendDir, pageName))
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

	A_SessionsMutex.RLock()
	defer A_SessionsMutex.RUnlock()

	u, ok := A_Sessions[sc.Value]
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
		"Title":    pageName + " - " + Hostname,
		"Config":   template.JS(data),
		"Hostname": Hostname,

		// page specific shit
		"BuildString": BuildString,
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
	fmt.Fprintln(w, Hostname)
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
	cmdStr, ok := REST_AllowedRootCommands[decoded["operation"]]
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
