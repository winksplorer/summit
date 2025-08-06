package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func C_SetValue(userId string, key string, val interface{}) error {
	A_SessionsMutex.Lock()
	defer A_SessionsMutex.Unlock()

	u, ok := A_Sessions[userId]
	if !ok {
		return fmt.Errorf("user not found: %s", userId)
	}

	return H_SetValue(u.config, key, val)
}

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
