package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"

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
	if !authenticated(w, r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	uc, err := r.Cookie("u")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	u, err := user.Lookup(uc.Value)
	if err != nil {
		fmt.Println("couldn't lookup username:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	uid, err := strconv.ParseInt(u.Uid, 10, 32)
	if err != nil {
		fmt.Println("couldn't parse uid:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	gid, err := strconv.ParseInt(u.Gid, 10, 32)
	if err != nil {
		fmt.Println("couldn't parse gid:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// upgrade to ws
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("couldn't upgrade to websockets:", err)
		return
	}
	defer conn.Close()

	// start pty
	cmd := exec.Command("bash")
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
	ptmx, err := pty.Start(cmd)
	if err != nil {
		fmt.Println("couldn't start pty:", err)
		return
	}
	defer ptmx.Close()

	// raw terminal
	oldState, err := term.MakeRaw(int(ptmx.Fd()))
	if err != nil {
		fmt.Println("couldn't set pty raw mode:", err)
		return
	}
	defer term.Restore(int(ptmx.Fd()), oldState)

	// pty -> websocket
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := ptmx.Read(buf)
			if err != nil {
				fmt.Println("couldn't read from pty:", err)
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, buf[:n]); err != nil {
				fmt.Println("couldn't send to websockets:", err)
				return
			}
		}
	}()

	// websocket -> pty
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("couldn't read from websockets:", err)
			break
		}

		// Try to parse resize message
		var resize termMessage
		if err := json.Unmarshal(msg, &resize); err == nil && resize.Type == "resize" {
			pty.Setsize(ptmx, &pty.Winsize{Cols: uint16(resize.Cols), Rows: uint16(resize.Rows)})
			continue
		}

		if _, err := ptmx.Write(msg); err != nil {
			fmt.Println("couldn't write to pty:", err)
			break
		}
	}
}
