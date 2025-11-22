package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const S_BlockDevDir = "/sys/dev/block"

type S_Device struct {
	id     string // x:y format
	parent string // id of parent device. empty if device is a parent
}

func S_GetDevices() ([]S_Device, error) {
	var devs []S_Device

	err := filepath.WalkDir(S_BlockDevDir, func(path string, d fs.DirEntry, err error) error {
		if path == S_BlockDevDir {
			return nil
		}

		realPath, err := filepath.EvalSymlinks(path)
		if err != nil {
			return err
		}

		id, err := os.ReadFile(realPath + "/dev")
		if err != nil {
			return err
		}

		_, err = os.Stat(realPath + "/partition")
		partition := !os.IsNotExist(err)

		var parent []byte
		if partition {
			parent, _ = os.ReadFile(filepath.Dir(realPath) + "/dev")
		} else {
			parent = []byte{}
		}

		devs = append(devs, S_Device{
			id:     strings.TrimSpace(string(id)),
			parent: strings.TrimSpace(string(parent)),
		})

		return nil
	})
	if err != nil {
		log.Printf("impossible to walk directories: %s", err)
		return nil, err
	}

	for _, v := range devs {
		log.Printf("ID: %s, P: %s", v.id, v.parent)
	}

	return devs, nil
}

func Comm_StorageGetdevs(data Comm_Message, keyCookie string) (any, error) {
	S_GetDevices()
	return map[string]any{}, nil
}
