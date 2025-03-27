package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
	"golang.org/x/term"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type termMessage struct {
	Type string `json:"type"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
}

func ptyHandler(w http.ResponseWriter, r *http.Request) {
	// upgrade to ws
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	// start pty
	cmd := exec.Command("bash")
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")
	ptmx, err := pty.Start(cmd)
	if err != nil {
		log.Println("Failed to start PTY:", err)
		return
	}
	defer ptmx.Close()

	// raw terminal
	oldState, err := term.MakeRaw(int(ptmx.Fd()))
	if err != nil {
		log.Println("Failed to set raw mode:", err)
		return
	}
	defer term.Restore(int(ptmx.Fd()), oldState)

	// pty -> websocket
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := ptmx.Read(buf)
			if err != nil {
				log.Println("PTY read error:", err)
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, buf[:n]); err != nil {
				log.Println("WebSocket write error:", err)
				return
			}
		}
	}()

	// websocket -> pty
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket read error:", err)
			break
		}

		// Try to parse resize message
		var resize termMessage
		if err := json.Unmarshal(msg, &resize); err == nil && resize.Type == "resize" {
			pty.Setsize(ptmx, &pty.Winsize{Cols: uint16(resize.Cols), Rows: uint16(resize.Rows)})
			continue
		}

		if _, err := ptmx.Write(msg); err != nil {
			log.Println("PTY write error:", err)
			break
		}
	}
}
