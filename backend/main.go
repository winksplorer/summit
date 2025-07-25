// summit backend/main.go - backend entry point + templating logic

package main

import (
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/NYTimes/gziphandler"
	"github.com/bddjr/hlfhr"
	"golang.org/x/net/http2"
)

var BuildDate string = "undefined"
var Version string = "undefined"
var frontendDir string = "/tmp/summit/frontend-dist"
var port string = ":7070"

var hostname string = "undefined"

func init() {
	// custom logging
	log.SetFlags(0)
	log.SetOutput(new(logWriter))
	log.Printf("summit v%s (built on %s).\n", Version, BuildDate)

	// select where the frontend is
	// SEA passes the first arg, so "summit dev" = "/tmp/summit/summit-server dev"
	if len(os.Args) > 1 && os.Args[1] == "dev" {
		frontendDir = "frontend"
	}

	// get the hostname
	localHostname, err := os.Hostname()
	if err != nil {
		log.Println("couldn't get hostname:", err)
	}

	hostname = localHostname
}

func main() {
	// configure endpoints
	http.HandleFunc("/", templater)
	http.HandleFunc("/api/login", loginHandler)
	http.HandleFunc("/api/logout", logoutHandler)
	http.HandleFunc("/api/get-hostname", getHostnameHandler)
	http.HandleFunc("/api/am-i-authed", amIAuthedHandler)
	http.HandleFunc("/api/sudo", sudoHandler)
	http.HandleFunc("/api/pty", ptyHandler)
	http.HandleFunc("/api/comm", commHandler)
	log.Println("successfully registered handlers")

	// configure server (hlfhr is used to redirect http to https)
	srv := hlfhr.New(&http.Server{
		Addr:    port,
		Handler: gziphandler.GzipHandler(http.DefaultServeMux),
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
		},
	})
	http2.ConfigureServer(srv.Server, &http2.Server{})

	srv.HttpOnHttpsPortErrorHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hlfhr.RedirectToHttps(w, r, 308)
	})

	// check if TLS shit exists
	_, err := os.Stat("/etc/ssl/private/summit.key")
	_, err2 := os.Stat("/etc/ssl/certs/summit.crt")
	if os.IsNotExist(err) || os.IsNotExist(err2) {
		log.Println("generating TLS certificates")
		// get hostname
		hostname, err := os.Hostname()
		if err != nil {
			log.Println("couldn't get hostname:", err)
			hostname = "undefined"
		}

		// generate tls cert & key
		if err := execCmd(
			"openssl",
			"req",
			"-x509",
			"-nodes",
			"-days", "365",
			"-newkey", "rsa:2048",
			"-keyout", "/etc/ssl/private/summit.key",
			"-out", "/etc/ssl/certs/summit.crt",
			"-subj", fmt.Sprintf("/C=US/ST=Washington/O=winksplorer & contributors/CN=summit (%s)", hostname),
		); err != nil {
			log.Println("couldn't generate certificate:", err)
			return
		}

		// change permissions
		if err := os.Chmod("/etc/ssl/private/summit.key", 0700); err != nil {
			log.Println("couldn't change private key permissions:", err)
			return
		}
	}

	log.Println("starting login auto-remove task")
	go removeOldSessions()

	log.Printf("summit on port %s\n", port)

	if err := srv.ListenAndServeTLS("/etc/ssl/certs/summit.crt", "/etc/ssl/private/summit.key"); err != nil {
		log.Println("error:", err)
	}
}

func templater(w http.ResponseWriter, r *http.Request) {
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
		ise(w, fmt.Sprintf("template parse error for %s", path), err)
		return
	}

	err = tmpl.ExecuteTemplate(w, pageName, map[string]interface{}{
		"Title": pageName + " - " + hostname,
	})
	if err != nil {
		ise(w, fmt.Sprintf("template exec error for %s", path), err)
		return
	}
}
