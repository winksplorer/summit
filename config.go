package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// copies user config template to `configFile`, and sets permissions
func C_Create(configFile string, uid uint32, gid uint32) error {
	log.Println("C_Create: Creating new configuration at", configFile+".")

	// copy the defaults
	defaultConfig, err := Frontend.ReadFile("frontend-dist/assets/defaultconfig.json")
	if err != nil {
		return err
	}

	if err := os.WriteFile(configFile, defaultConfig, 0664); err != nil {
		return err
	}

	// set permissions
	if err := os.Chmod(configFile, 0600); err != nil {
		return err
	}

	if err := os.Chown(configFile, int(uid), int(gid)); err != nil {
		return err
	}

	return nil
}

// sets a user config value
func C_SetValue(userId string, key string, val any) error {
	A_SessionsMutex.Lock()
	defer A_SessionsMutex.Unlock()

	u, ok := A_Sessions[userId]
	if !ok {
		return fmt.Errorf("user not found: %s", userId)
	}

	return IT_Set(u.config, key, val)
}

// saves user config
func C_Save(userId string) error {
	A_SessionsMutex.RLock()
	defer A_SessionsMutex.RUnlock()

	u, ok := A_Sessions[userId]
	if !ok {
		return fmt.Errorf("user not found: %s", userId)
	}

	data, err := json.MarshalIndent(u.config, "", "  ")
	if err != nil {
		return fmt.Errorf("config serialization error: %s", err)
	}

	if err := os.WriteFile(u.configFile, data, 0600); err != nil {
		return fmt.Errorf("config write error: %s", err)
	}

	return nil
}
