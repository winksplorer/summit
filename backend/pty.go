// summit backend/pty.go - handles websocket terminal

package main

import (
	"fmt"
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

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

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
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("couldn't upgrade to websockets:", err)
		return
	}
	defer conn.Close()

	authsMu.RLock()
	u := auths[sc.Value]
	authsMu.RUnlock()

	// get login shell
	out, err := exec.Command("getent", "passwd", strconv.FormatInt(int64(u.uid), 10)).Output()
	shell := "bash"
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
		fmt.Sprintf("HOME=%s", u.homedir),
		fmt.Sprintf("USER=%s", u.user),
		fmt.Sprintf("LOGNAME=%s", u.user),
		fmt.Sprintf("SHELL=%s", shell),
		"TERM=xterm-256color",
	)

	cmd.Dir = u.homedir

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

		// try to parse resize message
		var decoded map[string]interface{}

		if err := msgpack.Unmarshal(msg, &decoded); err == nil && decoded["type"] == "resize" {
			// resize pty
			if err := pty.Setsize(ptmx, &pty.Winsize{Cols: H_AsUint16(decoded["cols"]), Rows: H_AsUint16(decoded["rows"])}); err != nil {
				log.Println("couldn't resize pty")
				ptmx.Close()
				break
			}

			// alert the process that the resizing happened
			err = cmd.Process.Signal(syscall.SIGWINCH)
			if err != nil {
				log.Println("couldn't alert pty shell process of a resizing")
				ptmx.Close()
				break
			}
			continue
		}

		if _, err := ptmx.Write(msg); err != nil {
			log.Println("couldn't write to pty:", err)
			ptmx.Close()
			break
		}
	}
}
