package main

import (
	"fmt"
	"net/http"
	"time"
)

type authedUser struct {
	id string
	ua string
}

var auths []authedUser

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

		if !authenticated(w, r) {
			id, err := randomBase64String(32)
			if err != nil {
				fmt.Println("error: generate login:", err)
				return
			}
			auths = append(auths, authedUser{id: id, ua: r.UserAgent()})

			http.SetCookie(w, &http.Cookie{
				Name:  "st",
				Value: id,
				//Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
				Expires:  time.Now().Add(4 * time.Hour),
			})

			fmt.Printf("added {%s,%s} to known authed users\n", id, r.UserAgent())
		}
		http.Redirect(w, r, "/term.html", http.StatusFound)
	}
}

// handles /api/logout
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if authenticated(w, r) {
			cookie, err := r.Cookie("st")
			if err != nil {
				fmt.Println("error: st disappeared:", err)
				http.Error(w, "value magically disappeared", http.StatusInternalServerError)
				return
			}

			auths = removeAuth(auths, authedUser{id: cookie.Value, ua: r.UserAgent()})
			deleteAuthCookie(w)
		}
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

// removes an authorized user
func removeAuth(auths []authedUser, userToRemove authedUser) []authedUser {
	for i, v := range auths {
		if v == userToRemove {
			fmt.Printf("removed {%s,%s} from known authed users\n", v.id, v.ua)
			return append(auths[:i], auths[i+1:]...)
		}
	}
	return auths
}

func deleteAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:  "st",
		Value: "",
		//	Path:    "/",
		Expires: time.Now().Add(-time.Hour),
		MaxAge:  -1,
	})
}

// checks if user is authenticated
func authenticated(w http.ResponseWriter, r *http.Request) bool {
	cookie, err := r.Cookie("st")
	if err != nil {
		return false
	}
	for _, v := range auths {
		if v.id == cookie.Value {
			if v.ua == r.UserAgent() {
				return true
			}
			fmt.Println("ALERT!! SOMEONE ATTEMPTED TO LOG IN WITH A VALID TOKEN BUT FROM A DIFFERENT BROWSER! BE WARY ABOUT POSSIBLE FUTURE ATTACKS! REVOKING TOKEN...")
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
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprintln(w, "/term.html")
			return
		}
		http.Error(w, "No, you're not authed.", http.StatusUnauthorized)
	}
}
