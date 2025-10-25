// summit backend/main.go - backend entry point + templating logic

package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	BuildDate   string = "undefined"
	Version     string = "undefined"
	BuildString string = "undefined"

	// FrontendDir string = "/tmp/summit/frontend-dist"
	Hostname string = "undefined"

	WS_Upgrader = websocket.Upgrader{}

	//go:embed frontend-dist/*
	Frontend      embed.FS
	FrontendCache = map[string][]byte{}
	FrontendMu    sync.RWMutex
	StartTime     = time.Now()
)

// entry point
func main() {
	// custom logging
	log.SetFlags(0)
	log.SetOutput(new(logWriter))

	BuildString = "summit v" + Version + " (built on " + BuildDate + ")"
	log.Println(BuildString)

	// get the hostname
	var err error
	Hostname, err = os.Hostname()
	if err != nil {
		log.Fatalf("os.Hostname: %s.", err)
	}

	// init global config
	if _, err := os.Stat(GC_Path); os.IsNotExist(err) {
		if err = GC_Create(); err != nil {
			log.Fatalf("GC_Create: %s.", err)
		}
	}

	if err := GC_Read(); err != nil {
		log.Fatalf("GC_Read: %s.", err)
	}

	// set port from config
	port := fmt.Sprintf(":%d", IT_MustNumber(GC_Config, "port", uint16(7070)))

	// call init functions
	REST_Init()
	go A_RemoveExpiredSessions()

	if err := TLS_Init(); err != nil {
		log.Fatalf("TLS_Init: %s.", err)
	}

	srv, err := HTTP_Init(port)
	if err != nil {
		log.Fatalf("HTTP_Init: %s.", err)
	}

	log.Println("Initialized summit on port", port)

	// start server
	if err := srv.ListenAndServeTLS("/etc/ssl/certs/summit.crt", "/etc/ssl/private/summit.key"); err != nil {
		log.Fatalf("srv.ListenAndServeTLS: %s.", err)
	}
}
