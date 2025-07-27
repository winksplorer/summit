package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func getConfigValue(userId string, key string) (interface{}, error) {
	authsMu.RLock()
	defer authsMu.RUnlock()

	u, ok := auths[userId]
	if !ok {
		return nil, fmt.Errorf("user not found: %s", userId)
	}

	return getValue(u.config, key)
}

func setConfigValue(userId string, key string, val interface{}) error {
	authsMu.Lock()
	defer authsMu.Unlock()

	u, ok := auths[userId]
	if !ok {
		return fmt.Errorf("user not found: %s", userId)
	}

	return setValue(u.config, key, val)
}

func saveConfig(userId string) error {
	authsMu.RLock()
	defer authsMu.RUnlock()

	u, ok := auths[userId]
	if !ok {
		return fmt.Errorf("user not found: %s", userId)
	}

	data, err := json.Marshal(u.config)
	if err != nil {
		return fmt.Errorf("config serialization error: %s", err)
	}

	if err := os.WriteFile(u.configFile, data, 0600); err != nil {
		return fmt.Errorf("config write error: %s", err)
	}

	return nil
}
