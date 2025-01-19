package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/msteinert/pam"
)

var ips []string

// authenticates user with pam
func PAMAuth(serviceName, userName, passwd string) error {
	t, err := pam.StartFunc(serviceName, userName, func(s pam.Style, msg string) (string, error) {
		switch s {
		case pam.PromptEchoOff:
			return passwd, nil
		case pam.PromptEchoOn, pam.ErrorMsg, pam.TextInfo:
			return "", nil
		}
		return "", errors.New("unrecognized pam message style")
	})

	if err != nil {
		return err
	}
	if err = t.Authenticate(0); err != nil {
		return err
	}

	return nil
}

// handles /api/login
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			fmt.Println("error: failed to parse login:", err)
			return
		}

		if err := PAMAuth("passwd", r.FormValue("username"), r.FormValue("password")); err != nil {
			http.Redirect(w, r, "/?auth=fail", http.StatusFound)
			return
		}
		if !authenticated(strings.Split(r.RemoteAddr, ":")[0]) {
			ips = append(ips, strings.Split(r.RemoteAddr, ":")[0])
			fmt.Printf("added %s to known ips\n", strings.Split(r.RemoteAddr, ":")[0])
		}
		http.Redirect(w, r, "/term.html", http.StatusFound)
	}
}

// handles /api/logout
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if authenticated(strings.Split(r.RemoteAddr, ":")[0]) {
			ips = remove(ips, strings.Split(r.RemoteAddr, ":")[0])
		}
		fmt.Printf("removed %s from known ips\n", strings.Split(r.RemoteAddr, ":")[0])
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

// checks if user is authenticated
func authenticated(ip string) bool {
	for _, v := range ips {
		if v == ip {
			return true
		}
	}
	return false
}

func amIAuthedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if authenticated(strings.Split(r.RemoteAddr, ":")[0]) {
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprintln(w, "/term.html")
			return
		}
		http.Error(w, "No, you're not authed.", http.StatusUnauthorized)
	}
}
