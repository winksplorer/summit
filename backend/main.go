package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
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
	http.HandleFunc("/api/server-euid", servereuidHandler)

	port := ":7070"
	fmt.Printf("summit on port %s\n", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Println("error:", err)
	}
}
