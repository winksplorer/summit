package main

import (
	"log"
	"net/http"
	"time"
)

type authedUser struct {
	id   string
	ua   string
	user string
}

var auths []authedUser

// handles /api/login
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
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
			auths = append(auths, authedUser{id: id, ua: r.UserAgent(), user: r.FormValue("username")})

			http.SetCookie(w, &http.Cookie{
				Name:     "s",
				Value:    id + r.FormValue("username"),
				SameSite: http.SameSiteStrictMode,
				Expires:  time.Now().Add(4 * time.Hour),
			})

			log.Printf("added {%s%s,%s} to known authed users\n", id, r.FormValue("username"), r.UserAgent())
		}
		http.Redirect(w, r, "/term.html", http.StatusFound)
	}
}

// handles /api/logout
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if authenticated(w, r) {
			s, err := r.Cookie("s")
			if err != nil {
				log.Println("error: session disappeared:", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			auths = removeAuth(auths, authedUser{id: s.Value[:32], ua: r.UserAgent(), user: s.Value[32:]})
			deleteAuthCookie(w)
		}
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

// removes an authorized user
func removeAuth(auths []authedUser, userToRemove authedUser) []authedUser {
	for i, v := range auths {
		if v == userToRemove {
			log.Printf("removed {%s%s,%s} from known authed users\n", v.id, v.user, v.ua)
			return append(auths[:i], auths[i+1:]...)
		}
	}
	return auths
}

func deleteAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:    "s",
		Value:   "",
		Expires: time.Now().Add(-time.Hour),
		MaxAge:  -1,
	})
}

// checks if user is authenticated
func authenticated(w http.ResponseWriter, r *http.Request) bool {
	s, err := r.Cookie("s")
	if err != nil {
		return false
	}

	for _, v := range auths {
		if v.id == s.Value[:32] {
			if v.ua == r.UserAgent() && v.user == s.Value[32:] {
				return true
			}
			log.Println("ALERT!! either someone tried to spoof their username, or they tried to login from a different browser using your token. the token has now been marked as invalid.")
			auths = removeAuth(auths, v)
			deleteAuthCookie(w)
			return false
		}
	}
	return false
}

func amIAuthedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if authenticated(w, r) {
			http.Error(w, "OK", http.StatusOK)
			return
		}
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}
}
