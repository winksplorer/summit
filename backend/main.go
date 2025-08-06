// summit backend/main.go - backend entry point + templating logic

package main

import (
	"fmt"
	"log"
	"os"
)

var BuildDate string = "undefined"
var Version string = "undefined"
var buildString string = "undefined"
var hostname string = "undefined"

// todo: make these a global config
var frontendDir string = "/tmp/summit/frontend-dist"
var port string = ":7070"
var allowedSudoCommands = map[string]string{
	"reboot":   "/sbin/reboot",
	"poweroff": "/sbin/poweroff",
}

func main() {
	// custom logging
	log.SetFlags(0)
	log.SetOutput(new(logWriter))

	buildString = fmt.Sprintf("summit v%s (built on %s)", Version, BuildDate)
	log.Println(buildString)

	// select where the frontend is. SEA will pass all args.
	if len(os.Args) > 1 && os.Args[1] == "dev" {
		frontendDir = "frontend"
	}

	// get the hostname
	var err error
	hostname, err = os.Hostname()
	if err != nil {
		log.Println("couldn't get hostname:", err)
	}

	// call init functions
	REST_Init()

	if err := TLS_Init(); err != nil {
		log.Fatalf("TLS_Init: %s.", err)
	}

	srv, err := HTTP_Init()
	if err != nil {
		log.Fatalf("HTTP_Init: %s.", err)
	}

	go A_RemoveExpiredSessions()

	log.Printf("Initialized summit on port %s.\n", port)

	// start server
	if err := srv.ListenAndServeTLS("/etc/ssl/certs/summit.crt", "/etc/ssl/private/summit.key"); err != nil {
		log.Println("error:", err)
	}
}
