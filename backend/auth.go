// summit backend/auth.go - handles authentication and sessions

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/user"
	"strconv"
	"sync"
	"time"
)

// the maximum time a session can last
const A_SessionExpireTime = 4 * time.Hour

// a user session
type A_Session struct {
	user       string         // username
	gid        uint32         // unix gid
	uid        uint32         // unix uid
	suppgids   []uint32       // supplementary unix gids
	configFile string         // configuration file path
	config     map[string]any // cached config data
	homedir    string         // home directory
	ua         string         // user agent
	ip         string         // ip
	expires    time.Time      // time when this user's login expires
}

var (
	// mutex for A_Sessions
	A_SessionsMutex sync.RWMutex

	// stores all valid sessions
	A_Sessions = make(map[string]A_Session)
)

// handles /api/login
func REST_Login(w http.ResponseWriter, r *http.Request) {
	// only allow post
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// get login data
	if err := r.ParseForm(); err != nil {
		H_ISE(w, "REST_Login: Couldn't parse login", err)
		return
	}

	// disallow root
	if r.FormValue("username") == "root" {
		http.Redirect(w, r, "/?err=root", http.StatusFound)
		return
	}

	// log in with PAM
	if err := H_PamAuth("passwd", r.FormValue("username"), r.FormValue("password")); err != nil {
		http.Redirect(w, r, "/?err=auth", http.StatusFound)
		return
	}

	if !A_Authenticated(w, r) {
		// generate id & expire time
		id, err := H_RandomBase64(32)
		if err != nil {
			H_ISE(w, "REST_Login: Couldn't generate login id", err)
			return
		}

		expires := time.Now().Add(A_SessionExpireTime)

		// lookup user in the system
		u, err := user.Lookup(r.FormValue("username"))
		if err != nil {
			H_ISE(w, "REST_Login: Couldn't lookup username", err)
			return
		}

		uid, gid, groups, err := A_GetUnixInfo(u)
		if err != nil {
			H_ISE(w, "A_GetUnixInfo", err)
			return
		}

		// config logic
		var configData map[string]any
		configFile := fmt.Sprintf("%s/.config/summit.json", u.HomeDir)

		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			if err = C_Create(configFile, uid, gid); err != nil {
				H_ISE(w, "REST_Login: Couldn't create configuration file", err)
				return
			}
		}

		config, err := os.ReadFile(configFile)
		if err != nil {
			H_ISE(w, "REST_Login: Couldn't read config file", err)
			return
		}

		err = json.Unmarshal(config, &configData)
		if err != nil {
			H_ISE(w, "REST_Login: Couldn't parse config file", err)
			return
		}

		// create auth, server-side
		A_SessionsMutex.Lock()

		A_Sessions[id] = A_Session{
			ua:         r.UserAgent(),
			user:       u.Username,
			uid:        uint32(uid),
			gid:        uint32(gid),
			suppgids:   groups,
			configFile: configFile,
			config:     configData,
			homedir:    u.HomeDir,
			ip:         H_ClientIP(r),
			expires:    expires,
		}

		A_SessionsMutex.Unlock()

		// create auth, browser-side
		http.SetCookie(w, &http.Cookie{
			Name:     "s",
			Value:    id,
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
			Expires:  expires,
			Secure:   true,
			HttpOnly: true,
		})

		log.Printf("REST_Login: Authenticated a client from %s\n", H_ClientIP(r))
	}
	http.Redirect(w, r, "/terminal.html", http.StatusFound)
}

// handles /api/logout
func REST_Logout(w http.ResponseWriter, r *http.Request) {
	// only allow get
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if A_Authenticated(w, r) {
		s, err := r.Cookie("s")
		if err != nil {
			H_ISE(w, "REST_Logout: Couldn't get session cookie", err)
			return
		}

		A_Remove(s.Value, w)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

// gets uid, gid, and associated groups from a user
func A_GetUnixInfo(u *user.User) (uint32, uint32, []uint32, error) {
	// get uid
	uid, err := strconv.ParseUint(u.Uid, 10, 32)
	if err != nil {
		return 0, 0, nil, err
	}

	// get gid
	gid, err := strconv.ParseUint(u.Gid, 10, 32)
	if err != nil {
		return 0, 0, nil, err
	}

	// get group ids
	stringGroups, err := u.GroupIds()
	if err != nil {
		return 0, 0, nil, err
	}

	var groups []uint32 = make([]uint32, len(stringGroups))

	// convert group ids to uint32, because somebody thought a fucking STRING is the best way.
	for i, gidStr := range stringGroups {
		gidInt, err := strconv.Atoi(gidStr)
		if err != nil {
			return 0, 0, nil, err
		}
		groups[i] = uint32(gidInt)
	}

	return uint32(uid), uint32(gid), groups, nil
}

// deletes authentication, both server and browser side.
func A_Remove(id string, w http.ResponseWriter) {
	// server-side
	A_SessionsMutex.Lock()
	defer A_SessionsMutex.Unlock()
	delete(A_Sessions, id)

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
func A_RemoveExpiredSessions() {
	log.Println("A_RemoveExpiredSessions: Start session auto-remove task.")
	for {
		time.Sleep(10 * time.Minute)
		removed := 0
		A_SessionsMutex.Lock()
		for id, v := range A_Sessions {
			if v.expires.Before(time.Now()) {
				delete(A_Sessions, id)
				removed++
			}
		}
		A_SessionsMutex.Unlock()

		if removed > 0 {
			log.Printf("A_RemoveExpiredSessions: Removed %d sessions.\n", removed)
		}
	}
}

// checks if user is authenticated
func A_Authenticated(w http.ResponseWriter, r *http.Request) bool {
	s, err := r.Cookie("s")
	if err != nil {
		return false
	}

	A_SessionsMutex.RLock()
	v, ok := A_Sessions[s.Value]
	A_SessionsMutex.RUnlock()

	if !ok || v.ua != r.UserAgent() || v.ip != H_ClientIP(r) || v.expires.Before(time.Now()) || len(s.Value) != 32 {
		A_Remove(s.Value, w)
		return false
	}
	return true
}
