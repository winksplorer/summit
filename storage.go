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
	ID       string   `msgpack:"id"`       // maj:min format
	Name     string   `msgpack:"name"`     // device name (nvme0n1, loop0, etc.)
	Parent   string   `msgpack:"parent"`   // id of parent device. empty if device is a parent
	Children []string `msgpack:"children"` // ids of child devices. empty if no children
	Size     uint64   `msgpack:"size"`     // device size
	Readonly bool     `msgpack:"ro"`       // readonly?
	Model    string   `msgpack:"model"`    // device's model string. only on some parents
	Serial   string   `msgpack:"serial"`   // device's serial. only on some parents
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
		dev.ID = filepath.Base(path)
		dev.Name = filepath.Base(realPath)

		// read size
		sizeVal, err := os.ReadFile(filepath.Join(realPath, "size"))
		if err != nil {
			return err
		}

		dev.Size, err = strconv.ParseUint(strings.TrimSpace(string(sizeVal)), 10, 64)
		if err != nil {
			return err
		}
		dev.Size *= 512

		// readonly?
		roVal, err := os.ReadFile(filepath.Join(realPath, "ro"))
		if err != nil {
			return err
		}

		if strings.TrimSpace(string(roVal)) == "1" {
			dev.Readonly = true
		}

		// get parent
		if _, err = os.Stat(filepath.Join(realPath, "partition")); !os.IsNotExist(err) {
			parentVal, _ := os.ReadFile(filepath.Join(filepath.Dir(realPath), "dev"))
			dev.Parent = strings.TrimSpace(string(parentVal))
		}

		// get model & serial
		if _, err = os.Stat(filepath.Join(realPath, "device")); !os.IsNotExist(err) {
			modelVal, _ := os.ReadFile(filepath.Join(filepath.Dir(realPath), "model"))
			dev.Model = strings.TrimSpace(string(modelVal))

			serialVal, _ := os.ReadFile(filepath.Join(filepath.Dir(realPath), "serial"))
			dev.Serial = strings.TrimSpace(string(serialVal))
		}

		devs[dev.ID] = dev
		return nil
	})
	if err != nil {
		log.Printf("cannot walk directories: %s", err)
		return nil, err
	}

	// set children
	for k, v := range devs {
		if v.Parent != "" {
			if dev, ok := devs[v.Parent]; ok {
				dev.Children = append(devs[v.Parent].Children, k)
				devs[v.Parent] = dev
			}
		}
	}

	// debug
	for k, v := range devs {
		if v.Parent != "" {
			continue
		}

		ro := ""
		if v.Readonly {
			ro = " READONLY"
		}

		log.Print(k + " [" + v.Name + "] (" + H_HumanReadable(v.Size) + ") \"" + v.Model + "\" {" + v.Serial + "}" + ro)
		for _, c := range v.Children {
			ro := ""
			if devs[c].Readonly {
				ro = " READONLY"
			}

			log.Print("- " + c + " [" + devs[c].Name + "] (" + H_HumanReadable(devs[c].Size) + ") \"" + devs[c].Model + "\" {" + devs[c].Serial + "}" + ro)
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
