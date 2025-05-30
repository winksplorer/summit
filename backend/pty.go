package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
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

	sc, err := r.Cookie("s")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	u, err := user.Lookup(sc.Value[32:])
	if err != nil {
		log.Println("couldn't lookup username:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	uid, err := strconv.ParseInt(u.Uid, 10, 32)
	if err != nil {
		log.Println("couldn't parse uid:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	gid, err := strconv.ParseInt(u.Gid, 10, 32)
	if err != nil {
		log.Println("couldn't parse gid:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// upgrade to ws
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("couldn't upgrade to websockets:", err)
		return
	}
	defer conn.Close()

	// get login shell
	out, err := exec.Command("getent", "passwd", strconv.FormatInt(uid, 10)).Output()
	shell := "bash"
	if err == nil {
		shell = strings.Split(strings.TrimSuffix(string(out), "\n"), ":")[6]
	}

	// prepare shell
	cmd := exec.Command(shell)
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
	cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=/home/%s", sc.Value[32:]))
	cmd.Env = append(cmd.Env, fmt.Sprintf("USER=%s", sc.Value[32:]))
	cmd.Env = append(cmd.Env, "TERM=xterm-256color")
	cmd.Dir = fmt.Sprintf("/home/%s", sc.Value[32:])

	// start shell with pty
	ptmx, err := pty.Start(cmd)
	if err != nil {
		log.Println("couldn't start pty:", err)
		return
	}
	defer ptmx.Close()

	// pty -> websocket
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := ptmx.Read(buf)
			if err != nil {
				log.Println("couldn't read from pty:", err)
				ptmx.Close()
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, buf[:n]); err != nil {
				log.Println("couldn't send to websockets:", err)
				ptmx.Close()
				return
			}
		}
	}()

	// websocket -> pty
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("couldn't read from websockets:", err)
			ptmx.Close()
			break
		}

		// Try to parse resize message
		var resize termMessage
		if err := json.Unmarshal(msg, &resize); err == nil && resize.Type == "resize" {
			pty.Setsize(ptmx, &pty.Winsize{Cols: uint16(resize.Cols), Rows: uint16(resize.Rows)})
			continue
		}

		if _, err := ptmx.Write(msg); err != nil {
			log.Println("couldn't write to pty:", err)
			ptmx.Close()
			break
		}
	}
}
