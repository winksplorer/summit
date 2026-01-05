package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// *** generic
type (
	U_UpgradablePackage struct {
		Name       string `msgpack:"name"`        // name (sshd, nginx, etc.)
		CurrentVer string `msgpack:"current_ver"` // current version
		NewVer     string `msgpack:"new_ver"`     // the version that the package will upgrade to
	}

	U_PackageManager interface {
		Version() string
		ListUpgradable() ([]U_UpgradablePackage, error)
		UpdateIndex() error
	}
)

var U_PkgMgr U_PackageManager

func U_Init() error {
	log.Println("U_Init: Init software update manager.")

	pkgMgr, err := U_DeterminePackageManager()
	if err != nil {
		return err
	}

	switch pkgMgr {
	case "apk":
		U_PkgMgr = &U_ApkManager{}
	case "apt":
		U_PkgMgr = &U_AptManager{}
	default:
		return fmt.Errorf("unknown init system: %s", pkgMgr)
	}

	return nil
}

func U_DeterminePackageManager() (string, error) {
	// apk
	if _, err := os.Stat("/etc/apk"); err == nil {
		return "apk", nil
	}

	// apt
	if _, err := os.Stat("/etc/apt"); err == nil {
		return "apt", nil
	}

	return "", fmt.Errorf("unable to determine package manager")
}

func Comm_UpdatesPkgmgr(data Comm_Message, keyCookie string) (any, error) {
	return U_PkgMgr.Version(), nil
}

func Comm_UpdatesList(data Comm_Message, keyCookie string) (any, error) {
	return U_PkgMgr.ListUpgradable()
}

func Comm_UpdatesUpdateindex(data Comm_Message, keyCookie string) (any, error) {
	return nil, U_PkgMgr.UpdateIndex()
}

// *** apk (alpine)
type U_ApkManager struct{}

func (mgr *U_ApkManager) Version() string {
	v, err := H_Execute(false, "apk", "--version")
	if err != nil {
		return "unknown"
	}

	return v
}

func (mgr *U_ApkManager) ListUpgradable() ([]U_UpgradablePackage, error) {
	// list of package names and versions
	pkgListRaw, err := H_Execute(true, "apk", "upgrade", "--simulate")
	if err != nil {
		return nil, err
	}

	var pkgs []U_UpgradablePackage
	for _, v := range strings.Split(strings.TrimSpace(pkgListRaw), "\n") {
		if v[:2] == "OK" {
			break
		}

		fields := strings.Fields(v)

		pkgs = append(pkgs, U_UpgradablePackage{
			Name:       fields[2],
			CurrentVer: fields[3][1:],
			NewVer:     fields[5][:len(fields[5])-1],
		})
	}

	return pkgs, nil
}

func (mgr *U_ApkManager) UpdateIndex() error {
	_, err := H_Execute(true, "apk", "update")
	return err
}

// *** apt (debian-based distros)
type U_AptManager struct{}

func (mgr *U_AptManager) Version() string {
	v, err := H_Execute(false, "apt", "--version")
	if err != nil {
		return "unknown"
	}

	return v
}

func (mgr *U_AptManager) ListUpgradable() ([]U_UpgradablePackage, error) {
	// list of package names and versions
	pkgListRaw, err := H_Execute(true, "apt", "list", "--upgradable")
	if err != nil {
		return nil, err
	}

	var pkgs []U_UpgradablePackage
	for _, v := range strings.Split(strings.TrimSpace(pkgListRaw), "\n") {
		if v == "" || v == "Listing..." || v[:7] == "WARNING" { // this seems a little sketchy
			continue
		}

		fields := strings.Fields(v)

		pkgs = append(pkgs, U_UpgradablePackage{
			Name:       strings.Split(fields[0], "/")[0],
			CurrentVer: fields[5][:len(fields[5])-1],
			NewVer:     fields[1],
		})
	}

	return pkgs, nil
}

func (mgr *U_AptManager) UpdateIndex() error {
	_, err := H_Execute(true, "apt", "update")
	return err
}
