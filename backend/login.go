// summit backend/login.go - handles login

package main

import (
	"log"
	"net/http"
	"sync"
	"time"
)

const authExpireTime = 4 * time.Hour

type authedUser struct {
	user    string
	ua      string
	ip      string
	expires time.Time
}

var (
	authsMu sync.RWMutex
	auths   = make(map[string]authedUser)
)

// handles /api/login
func loginHandler(w http.ResponseWriter, r *http.Request) {
	// only allow post
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// get login data
	if err := r.ParseForm(); err != nil {
		log.Println("error: failed to parse login:", err)
		return
	}

	// disallow root
	if r.FormValue("username") == "root" {
		http.Redirect(w, r, "/?auth=rootfail", http.StatusFound)
		return
	}

	// log in with PAM
	if err := pamAuth("passwd", r.FormValue("username"), r.FormValue("password")); err != nil {
		http.Redirect(w, r, "/?auth=fail", http.StatusFound)
		return
	}

	if !authenticated(w, r) {
		// generate id & expire time
		id, err := randomBase64String(32)
		if err != nil {
			log.Println("error: generate login:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		expires := time.Now().Add(authExpireTime)

		// create auth, server-side
		authsMu.Lock()
		auths[id] = authedUser{
			ua:      r.UserAgent(),
			user:    r.FormValue("username"),
			ip:      clientIP(r),
			expires: expires,
		}
		authsMu.Unlock()

		// create auth, browser-side
		http.SetCookie(w, &http.Cookie{
			Name:     "s",
			Value:    id,
			SameSite: http.SameSiteStrictMode,
			Expires:  expires,
		})

		log.Printf("added a client from %s to known authed users\n", clientIP(r))
	}
	http.Redirect(w, r, "/terminal.html", http.StatusFound)
}

// handles /api/logout
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	// only allow get
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if authenticated(w, r) {
		s, err := r.Cookie("s")
		if err != nil {
			log.Println("error: session disappeared:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		deleteAuth(s.Value, w)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

// deletes authentication, both server and browser side.
func deleteAuth(id string, w http.ResponseWriter) {
	// server-side
	authsMu.Lock()
	delete(auths, id)
	authsMu.Unlock()

	// browser-side
	http.SetCookie(w, &http.Cookie{
		Name:     "s",
		Value:    "",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		SameSite: http.SameSiteStrictMode,
	})
}

// deletes expired sessions every 10 minutes
func removeOldSessions() {
	for {
		time.Sleep(10 * time.Minute)
		authsMu.Lock()
		for id, v := range auths {
			if v.expires.Before(time.Now()) {
				delete(auths, id)
			}
		}
		authsMu.Unlock()
	}
}

// checks if user is authenticated
func authenticated(w http.ResponseWriter, r *http.Request) bool {
	s, err := r.Cookie("s")
	if err != nil {
		return false
	}

	authsMu.RLock()
	v, ok := auths[s.Value]
	authsMu.RUnlock()

	if !ok || !userAgentMatches(v.ua, r.UserAgent()) || v.ip != clientIP(r) || v.expires.Before(time.Now()) || len(s.Value) != 32 {
		deleteAuth(s.Value, w)
		return false
	}
	return true
}

// http wrapper for authenticated(w,r). handles /api/am-i-authed.
func amIAuthedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if authenticated(w, r) {
		http.Error(w, "OK", http.StatusOK)
		return
	}

	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}
