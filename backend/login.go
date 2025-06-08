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
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Println("error: failed to parse login:", err)
		return
	}

	if r.FormValue("username") == "root" {
		http.Redirect(w, r, "/?auth=rootfail", http.StatusFound)
		return
	}

	if err := pamAuth("passwd", r.FormValue("username"), r.FormValue("password")); err != nil {
		http.Redirect(w, r, "/?auth=fail", http.StatusFound)
		return
	}

	if !authenticated(w, r) {
		id, err := randomBase64String(32)
		if err != nil {
			log.Println("error: generate login:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		expires := time.Now().Add(authExpireTime)

		authsMu.Lock()
		auths[id] = authedUser{
			ua:      r.UserAgent(),
			user:    r.FormValue("username"),
			ip:      clientIP(r),
			expires: expires,
		}
		authsMu.Unlock()

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

		authsMu.Lock()
		delete(auths, s.Value)
		authsMu.Unlock()

		deleteAuthCookie(w)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

// deletes s.
func deleteAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "s",
		Value:    "",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		SameSite: http.SameSiteStrictMode,
	})
}

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
		authsMu.Lock()
		delete(auths, s.Value)
		authsMu.Unlock()
		deleteAuthCookie(w)
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
