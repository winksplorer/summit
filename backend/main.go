package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./frontend"))))
	http.HandleFunc("/api/login", loginHandler)
	http.HandleFunc("/api/logout", logoutHandler)
	http.HandleFunc("/api/get-hostname", getHostnameHandler)
	http.HandleFunc("/api/stats", statsHandler)
	http.HandleFunc("/api/am-i-authed", amIAuthedHandler)

	port := ":7070"
	fmt.Printf("summit on port %s\n", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Println("error:", err)
	}
}
