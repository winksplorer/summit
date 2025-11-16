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

type (
	REST_TemplateData struct {
		Title       string
		Config      template.JS
		Hostname    string
		BuildString string
	}

	REST_RootRequest struct {
		Password  string
		Operation string
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

	pageName := strings.TrimSuffix(path, ".html")

	// template together base + the page
	var (
		tmpl *template.Template
		err  error
	)

	if G_FrontendOverride == "" {
		tmpl, err = template.ParseFS(G_Frontend, "frontend-dist/template/base.html", "frontend-dist/template/"+pageName+".html")
	} else {
		tmpl, err = template.ParseFiles(G_FrontendOverride+"/template/base.html", G_FrontendOverride+"/template/"+pageName+".html")
	}

	if err != nil && (strings.Contains(err.Error(), "pattern matches no files") || strings.Contains(err.Error(), "no such file or directory")) {
		// directly serve the file without any templating
		HTTP_ServeStatic(w, r, path)
		return
	} else if err != nil {
		H_ISE(w, "REST_Origin: Template parse error for "+path, err)
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
		H_ISE(w, "REST_Origin: Couldn't represent config as JSON", err)
		return
	}

	err = tmpl.ExecuteTemplate(w, pageName, REST_TemplateData{
		Title:    pageName + " - " + G_Hostname,
		Config:   template.JS(data),
		Hostname: G_Hostname,

		// page specific shit
		BuildString: G_BuildString,
	})
	if err != nil {
		H_ISE(w, "REST_Origin: Template exec error for "+path, err)
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
	fmt.Fprintln(w, G_Hostname)
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
		H_ISE(w, "REST_SUID: Couldn't read request data", err)
		return
	}
	defer r.Body.Close()

	// parse
	var rootReq REST_RootRequest
	if err := json.Unmarshal(body, &rootReq); err != nil {
		log.Println("REST_SUID: Couldn't parse request data:", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// authenticate as root with PAM
	if err := H_PamAuth("passwd", "root", rootReq.Password); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// see what command we need
	cmdStr, ok := REST_AllowedRootCommands[rootReq.Operation]
	if !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// execute command
	cmd := exec.Command(cmdStr)
	err = cmd.Run()
	if err != nil {
		H_ISE(w, "REST_SUID: Couldn't execute requested action", err)
		return
	}

	// we good
	http.Error(w, "OK", http.StatusOK)
}
