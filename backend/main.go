package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"

	"github.com/bddjr/hlfhr"
	"golang.org/x/net/http2"
)

var BuildDate string = "undefined"
var Version string = "undefined"
var Edition string = "undefined"

func main() {
	port := ":7070"

	var frontendDir string
	if len(os.Args) > 1 && os.Args[1] == "dev" {
		frontendDir = "./frontend"
	} else {
		frontendDir = "/tmp/summit/frontend"
	}

	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(frontendDir))))
	http.HandleFunc("/api/login", loginHandler)
	http.HandleFunc("/api/logout", logoutHandler)
	http.HandleFunc("/api/get-hostname", getHostnameHandler)
	http.HandleFunc("/api/stats", statsHandler)
	http.HandleFunc("/api/am-i-authed", amIAuthedHandler)
	http.HandleFunc("/api/server-pages", serverPagesHandler)
	http.HandleFunc("/api/buildstring", buildstringHandler)
	http.HandleFunc("/api/sudo", sudoHandler)
	http.HandleFunc("/api/pty", ptyHandler)

	srv := hlfhr.New(&http.Server{
		Addr: port,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
		},
	})
	http2.ConfigureServer(srv.Server, &http2.Server{})

	srv.HttpOnHttpsPortErrorHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hlfhr.RedirectToHttps(w, r, 308)
	})

	log.SetFlags(0)
	log.SetOutput(new(logWriter))

	log.Printf("summit on port %s\n", port)

	if err := srv.ListenAndServeTLS("/tmp/summit/summit.crt", "/tmp/summit/summit.key"); err != nil {
		log.Println("error:", err)
	}
}
