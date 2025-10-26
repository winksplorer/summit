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

const G_LicensingMsg = "summit is licensed under Apache-2.0 and includes third-party components under MIT and OFL licenses. Copyright (c) 2025 winksplorer et al."

var (
	G_BuildDate   string
	G_Version     string
	G_BuildString string

	G_Hostname   string
	G_WSUpgrader = websocket.Upgrader{}
	G_StartTime  = time.Now()

	//go:embed frontend-dist/*
	G_Frontend      embed.FS
	G_FrontendCache = map[string][]byte{}
	G_FrontendMu    sync.RWMutex
)

// entry point
func main() {
	// custom logging
	log.SetFlags(0)
	log.SetOutput(new(LogWriter))

	G_BuildString = "summit v" + G_Version + " (built on " + G_BuildDate + ")"
	log.Println(G_BuildString)

	log.Println(G_LicensingMsg)

	if os.Geteuid() != 0 {
		log.Fatalln("Root permissions are required for summit to work correctly.")
	}

	// get the hostname
	var err error
	G_Hostname, err = os.Hostname()
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
	GC_ConfigMu.RLock()
	port := fmt.Sprintf(":%d", IT_MustNumber(GC_Config, "port", uint16(7070)))
	GC_ConfigMu.RUnlock()

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
