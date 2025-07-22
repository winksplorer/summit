package main

import (
	"encoding/json"
	"log"
	"os"
)

// creates the config file for a user
func createConfig(userId string) error {
	authsMu.RLock()
	u := auths[userId]
	authsMu.RUnlock()

	// sets up "data"
	data, err := json.Marshal(map[string]interface{}{})
	if err != nil {
		log.Println("couldn't create initial config data:", err)
		return err
	}

	// creates & writes file
	if err := os.WriteFile(u.config, data, 0700); err != nil {
		log.Println("couldn't write config file:", err)
		return err
	}

	// sets permissions of file
	if err := os.Chown(u.config, int(u.uid), int(u.gid)); err != nil {
		log.Println("couldn't set permissions of config file:", err)
		return err
	}

	return nil
}
