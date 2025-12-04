package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const S_BlockDevDir = "/sys/dev/block"

type S_Device struct {
	id       string   // maj:min format
	name     string   // device name (nvme0n1, loop0, etc.)
	parent   string   // id of parent device. empty if device is a parent
	children []string // ids of child devices. empty if no children
	size     uint64   // device size
	ro       bool     // readonly?
	model    string   // device's model string. only on some parents
	serial   string   // device's serial. only on some parents
}

func S_GetDevices() (map[string]S_Device, error) {
	devs := make(map[string]S_Device)

	err := filepath.WalkDir(S_BlockDevDir, func(path string, d fs.DirEntry, err error) error {
		if path == S_BlockDevDir || err != nil {
			return nil
		}

		dev := S_Device{}

		// get real path
		realPath, err := filepath.EvalSymlinks(path)
		if err != nil {
			return err
		}

		// get id and name
		dev.id = filepath.Base(path)
		dev.name = filepath.Base(realPath)

		// read size
		sizeVal, err := os.ReadFile(filepath.Join(realPath, "size"))
		if err != nil {
			return err
		}

		dev.size, err = strconv.ParseUint(strings.TrimSpace(string(sizeVal)), 10, 64)
		if err != nil {
			return err
		}
		dev.size *= 512

		// readonly?
		roVal, err := os.ReadFile(filepath.Join(realPath, "ro"))
		if err != nil {
			return err
		}

		if strings.TrimSpace(string(roVal)) == "1" {
			dev.ro = true
		}

		// get parent
		if _, err = os.Stat(filepath.Join(realPath, "partition")); !os.IsNotExist(err) {
			parentVal, _ := os.ReadFile(filepath.Join(filepath.Dir(realPath), "dev"))
			dev.parent = strings.TrimSpace(string(parentVal))
		}

		// get model & serial
		if _, err = os.Stat(filepath.Join(realPath, "device")); !os.IsNotExist(err) {
			modelVal, _ := os.ReadFile(filepath.Join(filepath.Dir(realPath), "model"))
			dev.model = strings.TrimSpace(string(modelVal))

			serialVal, _ := os.ReadFile(filepath.Join(filepath.Dir(realPath), "serial"))
			dev.serial = strings.TrimSpace(string(serialVal))
		}

		devs[dev.id] = dev
		return nil
	})
	if err != nil {
		log.Printf("cannot walk directories: %s", err)
		return nil, err
	}

	// set children
	for k, v := range devs {
		if v.parent != "" {
			if dev, ok := devs[v.parent]; ok {
				dev.children = append(devs[v.parent].children, k)
				devs[v.parent] = dev
			}
		}
	}

	// debug
	for k, v := range devs {
		if v.parent != "" {
			continue
		}

		ro := ""
		if v.ro {
			ro = " READONLY"
		}

		log.Print(k + " [" + v.name + "] (" + H_HumanReadable(v.size) + ") \"" + v.model + "\" {" + v.serial + "}" + ro)
		for _, c := range v.children {
			ro := ""
			if devs[c].ro {
				ro = " READONLY"
			}

			log.Print("- " + c + " [" + devs[c].name + "] (" + H_HumanReadable(devs[c].size) + ") \"" + devs[c].model + "\" {" + devs[c].serial + "}" + ro)
		}
	}

	return devs, nil
}

func Comm_StorageGetdevs(data Comm_Message, keyCookie string) (any, error) {
	devs, err := S_GetDevices()
	if err != nil {
		return nil, err
	}

	return devs, nil
}
