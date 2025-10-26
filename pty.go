// summit backend/pty.go - handles websocket terminal

package main

import (
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
	"github.com/vmihailenco/msgpack/v5"
)

// msgpack request made when terminal is resized
type PTY_ResizeRequest struct {
	Type string
	Rows uint16
	Cols uint16
}

// websocket terminal. handles /api/pty
func REST_Pty(w http.ResponseWriter, r *http.Request) {
	if !A_Authenticated(w, r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sc, err := r.Cookie("s")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// upgrade to ws
	conn, err := G_WSUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("REST_Pty: Couldn't upgrade to WebSocket:", err)
		return
	}
	defer conn.Close()

	A_SessionsMutex.RLock()
	u := A_Sessions[sc.Value]
	A_SessionsMutex.RUnlock()

	// get login shell
	out, err := exec.Command("getent", "passwd", strconv.FormatInt(int64(u.uid), 10)).Output()
	shell := "/usr/bin/sh"
	if err == nil {
		shell = strings.Split(strings.TrimSuffix(string(out), "\n"), ":")[6]
	}

	// prepare shell
	cmd := exec.Command(shell)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid:    u.uid,
			Gid:    u.gid,
			Groups: u.suppgids,
		},
	}

	// env variables & directory
	cmd.Env = append(cmd.Env,
		"HOME="+u.homedir,
		"USER="+u.user,
		"LOGNAME="+u.user,
		"SHELL="+shell,
		"TERM=xterm-256color",
	)

	cmd.Dir = u.homedir

	// start shell with pty
	ptmx, err := pty.Start(cmd)
	if err != nil {
		log.Println("REST_Pty: Couldn't start PTY:", err)
		return
	}
	defer ptmx.Close()

	// pty -> websocket
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := ptmx.Read(buf)
			if err != nil {
				log.Println("REST_Pty: Couldn't read from PTY:", err)
				ptmx.Close()
				conn.Close()
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, buf[:n]); err != nil {
				log.Println("REST_Pty: Couldn't send to WebSocket:", err)
				ptmx.Close()
				conn.Close()
				return
			}
		}
	}()

	// websocket -> pty
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("REST_Pty: Couldn't read from WebSocket:", err)
			ptmx.Close()
			break
		}

		// try to parse resize message
		var resizeReq PTY_ResizeRequest
		if err = msgpack.Unmarshal(msg, &resizeReq); err == nil && resizeReq.Type == "resize" {
			// resize pty
			if err := pty.Setsize(ptmx, &pty.Winsize{Cols: resizeReq.Cols, Rows: resizeReq.Rows}); err != nil {
				log.Println("REST_Pty: Couldn't resize PTY:", err)
				ptmx.Close()
				break
			}

			// alert the process that the resizing happened
			err = cmd.Process.Signal(syscall.SIGWINCH)
			if err != nil {
				log.Println("REST_Pty: Couldn't alert PTY shell process of a resizing:", err)
				ptmx.Close()
				break
			}
			continue
		}

		if _, err := ptmx.Write(msg); err != nil {
			log.Println("REST_Pty: Couldn't write to PTY:", err)
			ptmx.Close()
			break
		}
	}
}
