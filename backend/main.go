// summit backend/main.go - backend entry point + templating logic

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

var (
	BuildDate   string = "undefined"
	Version     string = "undefined"
	BuildString string = "undefined"
	Hostname    string = "undefined"
	WS_Upgrader        = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	FrontendDir string = "/tmp/summit/frontend-dist"
	Port        string = ":7070"
)

func main() {
	// custom logging
	log.SetFlags(0)
	log.SetOutput(new(logWriter))

	BuildString = fmt.Sprintf("summit v%s (built on %s)", Version, BuildDate)
	log.Println(BuildString)

	// select where the frontend is. SEA will pass all args.
	if len(os.Args) > 1 && os.Args[1] == "dev" {
		log.Println("Using ./frontend for frontend directory.")
		FrontendDir = "frontend"
	}

	// get the hostname
	var err error
	Hostname, err = os.Hostname()
	if err != nil {
		log.Fatalf("os.Hostname: %s.", err)
	}

	// call init functions
	REST_Init()
	go A_RemoveExpiredSessions()

	if err := TLS_Init(); err != nil {
		log.Fatalf("TLS_Init: %s.", err)
	}

	srv, err := HTTP_Init()
	if err != nil {
		log.Fatalf("HTTP_Init: %s.", err)
	}

	log.Printf("Initialized summit on port %s.\n", Port)

	// start server
	if err := srv.ListenAndServeTLS("/etc/ssl/certs/summit.crt", "/etc/ssl/private/summit.key"); err != nil {
		log.Fatalf("srv.ListenAndServeTLS: %s.", err)
	}
}
