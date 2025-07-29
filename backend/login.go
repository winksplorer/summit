// summit backend/login.go - handles login

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

const authExpireTime = 4 * time.Hour

type authedUser struct {
	user       string                 // username
	gid        uint32                 // unix gid
	uid        uint32                 // unix uid
	suppgids   []uint32               // supplementary unix gids
	configFile string                 // configuration file path
	config     map[string]interface{} // cached config data
	homedir    string                 // home directory
	ua         string                 // user agent
	ip         string                 // ip
	expires    time.Time              // time when this user's login expires
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
		ise(w, "couldn't parse login", err)
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
			ise(w, "couldn't generate login id", err)
			return
		}

		expires := time.Now().Add(authExpireTime)

		// lookup user in the system
		u, err := user.Lookup(r.FormValue("username"))
		if err != nil {
			ise(w, "couldn't lookup username", err)
			return
		}

		// get uid
		uid, err := strconv.ParseUint(u.Uid, 10, 32)
		if err != nil {
			ise(w, "couldn't parse uid", err)
			return
		}

		// get gid
		gid, err := strconv.ParseUint(u.Gid, 10, 32)
		if err != nil {
			ise(w, "couldn't parse gid", err)
			return
		}

		// get group ids
		stringGroups, err := u.GroupIds()
		if err != nil {
			ise(w, "couldn't get group ids", err)
			return
		}

		var groups []uint32 = make([]uint32, len(stringGroups))

		for i, gidStr := range stringGroups {
			gidInt, err := strconv.Atoi(gidStr)
			if err != nil {
				ise(w, "couldn't convert group ids", err)
				return
			}
			groups[i] = uint32(gidInt)
		}

		// config logic
		var configData map[string]interface{}
		configFile := fmt.Sprintf("%s/.config/summit.json", u.HomeDir)

		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			// copy the defaults
			if err = copyFile(fmt.Sprintf("%s/assets/defaultconfig.json", frontendDir), configFile); err != nil {
				ise(w, "couldn't create configuration file", err)
				return
			}

			// set permissions
			if err = os.Chmod(configFile, 0600); err != nil {
				ise(w, "couldn't set permissions of config file (chmod)", err)
				return
			}

			if err = os.Chown(configFile, int(uid), int(gid)); err != nil {
				ise(w, "couldn't set permissions of config file (chown)", err)
				return
			}
		}

		config, err := os.ReadFile(configFile)
		if err != nil {
			ise(w, "couldn't read config file", err)
			return
		}

		err = json.Unmarshal(config, &configData)
		if err != nil {
			ise(w, "couldn't parse config file", err)
			return
		}

		// create auth, server-side
		authsMu.Lock()
		defer authsMu.Unlock()

		auths[id] = authedUser{
			ua:         r.UserAgent(),
			user:       u.Username,
			uid:        uint32(uid),
			gid:        uint32(gid),
			suppgids:   groups,
			configFile: configFile,
			config:     configData,
			homedir:    u.HomeDir,
			ip:         clientIP(r),
			expires:    expires,
		}

		// create auth, browser-side
		http.SetCookie(w, &http.Cookie{
			Name:     "s",
			Value:    id,
			Path:     "/",
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
			ise(w, "couldn't get session cookie", err)
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
	defer authsMu.Unlock()
	delete(auths, id)

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
