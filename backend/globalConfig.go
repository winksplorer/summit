package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// the path to the global config
const GC_Path string = "/etc/summit.json"

// cached global config
var GC_Config map[string]interface{}

// copies global config template to GC_Path
func GC_Create() error {
	log.Printf("GC_Create: Creating new configuration at %s.", GC_Path)

	// copy the defaults
	if err := H_CopyFile(fmt.Sprintf("%s/assets/defaultgc.json", FrontendDir), GC_Path); err != nil {
		return err
	}

	// set permissions
	if err := os.Chmod(GC_Path, 0664); err != nil {
		return err
	}

	return nil
}

// sets a value in GC_Config
func GC_SetValue(key string, val interface{}) error {
	return IT_Set(GC_Config, key, val)
}

// fills in GC_Config with the actual data from GC_Path
func GC_Read() error {
	config, err := os.ReadFile(GC_Path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(config, &GC_Config)
	if err != nil {
		return err
	}

	return nil
}

// saves GC_Config to GC_path
func GC_Save() error {
	data, err := json.MarshalIndent(GC_Config, "", "  ")
	if err != nil {
		return fmt.Errorf("config serialization error: %s", err)
	}

	if err := os.WriteFile(GC_Path, data, 0664); err != nil {
		return fmt.Errorf("config write error: %s", err)
	}

	return nil
}
