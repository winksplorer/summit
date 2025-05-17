package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/bddjr/hlfhr"
	"golang.org/x/net/http2"
)

var BuildDate string = "undefined"
var Version string = "undefined"

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

		// generate tls cert
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

	log.Printf("summit on port %s\n", port)

	if err := srv.ListenAndServeTLS("/etc/ssl/certs/summit.crt", "/etc/ssl/private/summit.key"); err != nil {
		log.Println("error:", err)
	}
}
