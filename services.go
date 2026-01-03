package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

// *** generic
type (
	S_Service struct {
		Name        string `msgpack:"name"`        // name (sshd, nginx, etc.)
		Description string `msgpack:"description"` // description
		Running     bool   `msgpack:"running"`     // service running?
		Enabled     bool   `msgpack:"enabled"`     // service enabled?
	}

	S_ServiceManager interface {
		Version() string
		ListServices() ([]S_Service, error)
		StartService(name string) error
		StopService(name string) error
		RestartService(name string) error
	}
)

var (
	S_SrvMgr     S_ServiceManager
	S_SrvMgrOnce sync.Once
	S_SrvMgrErr  error
)

func S_DetermineInitSystem() (string, error) {
	comm, err := os.ReadFile("/proc/1/comm")
	if err != nil {
		return "", err
	}

	// check command name
	switch strings.TrimSpace(string(comm)) {
	case "systemd":
		return "systemd", nil
	case "openrc-init":
		return "openrc", nil
	case "runit":
		return "runit", nil
	}

	// not sure? check for /run/openrc
	if _, err = os.Stat("/run/openrc"); err == nil {
		return "openrc", nil
	}

	return "", fmt.Errorf("unable to determine init system: %s", string(comm))
}

func S_GetServiceManager() (S_ServiceManager, error) {
	S_SrvMgrOnce.Do(func() {
		initSystem, err := S_DetermineInitSystem()
		if err != nil {
			S_SrvMgrErr = err
			return
		}

		switch initSystem {
		case "systemd":
			S_SrvMgr = &S_SystemdManager{}
		case "openrc":
			S_SrvMgr = &S_OpenRCManager{}
		default:
			S_SrvMgrErr = fmt.Errorf("unknown init system: %s", initSystem)
		}
	})

	return S_SrvMgr, S_SrvMgrErr
}

func Comm_SrvInitver(data Comm_Message, keyCookie string) (any, error) {
	mgr, err := S_GetServiceManager()
	if err != nil {
		return nil, err
	}

	return mgr.Version(), nil
}

func Comm_SrvList(data Comm_Message, keyCookie string) (any, error) {
	mgr, err := S_GetServiceManager()
	if err != nil {
		return nil, err
	}

	return mgr.ListServices()
}

func Comm_SrvStart(data Comm_Message, keyCookie string) (any, error) {
	mgr, err := S_GetServiceManager()
	if err != nil {
		return nil, err
	}

	service, ok := data.Data.(string)
	if !ok {
		return nil, err
	}

	return nil, mgr.StartService(service)
}

func Comm_SrvStop(data Comm_Message, keyCookie string) (any, error) {
	mgr, err := S_GetServiceManager()
	if err != nil {
		return nil, err
	}

	service, ok := data.Data.(string)
	if !ok {
		return nil, err
	}

	return nil, mgr.StopService(service)
}

func Comm_SrvRestart(data Comm_Message, keyCookie string) (any, error) {
	mgr, err := S_GetServiceManager()
	if err != nil {
		return nil, err
	}

	service, ok := data.Data.(string)
	if !ok {
		return nil, err
	}

	return nil, mgr.RestartService(service)
}

// *** openrc
type S_OpenRCManager struct{}

func (mgr *S_OpenRCManager) Version() string {
	v, err := H_Execute("openrc", "-V")
	if err != nil {
		return "unknown"
	}

	return v
}

func (mgr *S_OpenRCManager) ListServices() ([]S_Service, error) {
	list, err := H_Execute("rc-service", "-l")
	if err != nil {
		return nil, err
	}

	srvList := strings.Split(strings.TrimSpace(list), "\n")
	var services []S_Service

	for _, srv := range srvList {
		info, err := H_Execute("rc-service", srv, "status", "describe")
		if err != nil {
			return nil, err
		}

		infoLines := strings.Split(strings.TrimSpace(info), "\n")

		services = append(services, S_Service{
			Name:        srv,
			Description: infoLines[1][3:],
			Running:     strings.Fields(infoLines[0])[2] == "started",
		})
	}

	return services, nil
}

func (mgr *S_OpenRCManager) StartService(name string) error {
	_, err := H_Execute("rc-service", name, "start")
	return err
}

func (mgr *S_OpenRCManager) StopService(name string) error {
	_, err := H_Execute("rc-service", name, "stop")
	return err
}

func (mgr *S_OpenRCManager) RestartService(name string) error {
	_, err := H_Execute("rc-service", name, "restart")
	return err
}

// *** systemd
type S_SystemdManager struct{}

func (mgr *S_SystemdManager) Version() string {
	v, err := H_Execute("systemctl", "--version")
	if err != nil {
		return "unknown"
	}

	return strings.Split(strings.TrimSpace(v), "\n")[0]
}

func (mgr *S_SystemdManager) ListServices() ([]S_Service, error) {
	list, err := H_Execute("systemctl", "list-units", "--type=service", "--all", "--no-pager", "--no-legend")
	if err != nil {
		return nil, err
	}

	srvList := strings.Split(strings.TrimSpace(list), "\n")
	var services []S_Service

	for _, srv := range srvList {
		if srv[:3] == "\u25cf" {
			srv = srv[3:]
		}

		fields := strings.Fields(srv)

		services = append(services, S_Service{
			Name:        fields[0],
			Description: strings.Join(fields[4:], " "),
			Running:     fields[3] == "running",
		})
	}

	return services, nil
}

func (mgr *S_SystemdManager) StartService(name string) error {
	_, err := H_Execute("systemctl", "start", name)
	return err
}

func (mgr *S_SystemdManager) StopService(name string) error {
	_, err := H_Execute("systemctl", "stop", name)
	return err
}

func (mgr *S_SystemdManager) RestartService(name string) error {
	_, err := H_Execute("systemctl", "restart", name)
	return err
}
