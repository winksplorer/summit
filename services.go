package main

import (
	"fmt"
	"log"
	"os"
	"slices"
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
		IsStub() bool
		Version() string
		ListServices() ([]S_Service, error)
		StartService(name string) error
		StopService(name string) error
		RestartService(name string) error
		EnableService(name string) error
		DisableService(name string) error
	}
)

var S_SrvMgr S_ServiceManager

func S_Init() error {
	log.Println("S_Init: Init service manager.")

	initSystem := S_DetermineInitSystem()

	switch initSystem {
	case "systemd":
		S_SrvMgr = &S_SystemdManager{}
	case "openrc":
		S_SrvMgr = &S_OpenRCManager{}
	default:
		log.Println("U_Init: Unknown init system, using stub.")
		S_SrvMgr = &S_StubManager{}
	}

	return nil
}

func S_DetermineInitSystem() string {
	comm, err := os.ReadFile("/proc/1/comm")
	if err != nil {
		return ""
	}

	// check command name
	switch strings.TrimSpace(string(comm)) {
	case "systemd":
		return "systemd"
	case "openrc-init":
		return "openrc"
	case "runit":
		return "runit"
	}

	// not sure? check for /run/openrc
	if _, err = os.Stat("/run/openrc"); err == nil {
		return "openrc"
	}

	return ""
}

func S_CommOp(data Comm_Message, f func(string) error) (any, error) {
	str, ok := data.Data.(string)
	if !ok {
		return nil, fmt.Errorf("data: not a string")
	}

	return nil, f(str)
}

func Comm_SrvInitver(data Comm_Message, keyCookie string) (any, error) {
	return S_SrvMgr.Version(), nil
}

func Comm_SrvList(data Comm_Message, keyCookie string) (any, error) {
	return S_SrvMgr.ListServices()
}

func Comm_SrvStart(data Comm_Message, keyCookie string) (any, error) {
	return S_CommOp(data, S_SrvMgr.StartService)
}

func Comm_SrvStop(data Comm_Message, keyCookie string) (any, error) {
	return S_CommOp(data, S_SrvMgr.StopService)
}

func Comm_SrvRestart(data Comm_Message, keyCookie string) (any, error) {
	return S_CommOp(data, S_SrvMgr.RestartService)
}

func Comm_SrvEnable(data Comm_Message, keyCookie string) (any, error) {
	return S_CommOp(data, S_SrvMgr.EnableService)
}

func Comm_SrvDisable(data Comm_Message, keyCookie string) (any, error) {
	return S_CommOp(data, S_SrvMgr.DisableService)
}

// *** stub
type S_StubManager struct{}

func (mgr *S_StubManager) IsStub() bool {
	return true
}

func (mgr *S_StubManager) Version() string {
	return "stub"
}

func (mgr *S_StubManager) ListServices() ([]S_Service, error) {
	return []S_Service{}, fmt.Errorf("using stub manager")
}

func (mgr *S_StubManager) StartService(name string) error {
	return fmt.Errorf("using stub manager")
}

func (mgr *S_StubManager) StopService(name string) error {
	return fmt.Errorf("using stub manager")
}

func (mgr *S_StubManager) RestartService(name string) error {
	return fmt.Errorf("using stub manager")
}

func (mgr *S_StubManager) EnableService(name string) error {
	return fmt.Errorf("using stub manager")
}

func (mgr *S_StubManager) DisableService(name string) error {
	return fmt.Errorf("using stub manager")
}

// *** openrc
type S_OpenRCManager struct{}

func (mgr *S_OpenRCManager) IsStub() bool {
	return false
}

func (mgr *S_OpenRCManager) Version() string {
	v, err := H_Execute(false, "openrc", "-V")
	if err != nil {
		return "unknown"
	}

	return v
}

func (mgr *S_OpenRCManager) ListServices() ([]S_Service, error) {
	// list of service names
	srvListRaw, err := H_Execute(false, "rc-service", "-l")
	if err != nil {
		return nil, err
	}

	// list of enabled services (and other bs we dont care about)
	enabledRaw, err := H_Execute(false, "rc-update", "show", "default")
	if err != nil {
		return nil, err
	}

	srvList := strings.Split(strings.TrimSpace(srvListRaw), "\n")

	// just get the names of enabled services
	var enabledList []string
	for _, line := range strings.Split(strings.TrimSpace(enabledRaw), "\n") {
		enabledList = append(enabledList, strings.Fields(line)[0])
	}

	// get description and status for each, parallel
	var services []S_Service
	srvChannel := make(chan S_Service, len(srvList))
	errChannel := make(chan error, 1)
	var wg sync.WaitGroup

	for _, srv := range srvList {
		wg.Add(1)
		go mgr.ServiceStatusAndDescription(srv, enabledList, srvChannel, errChannel, &wg)
	}

	wg.Wait()
	close(srvChannel)
	close(errChannel)

	// error present? return it
	if err := <-errChannel; err != nil {
		return nil, err
	}

	// add all the services
	for service := range srvChannel {
		services = append(services, service)
	}

	return services, nil
}

func (mgr *S_OpenRCManager) ServiceStatusAndDescription(name string, enabledList []string, srvChannel chan S_Service, errChannel chan error, wg *sync.WaitGroup) {
	defer wg.Done()

	// get data
	info, err := H_Execute(false, "rc-service", name, "status", "describe")
	if err != nil && !strings.Contains(err.Error(), "exit status 3") {
		select {
		case errChannel <- err:
		default:
		}
		return
	}

	// create S_Service
	infoLines := strings.Split(strings.TrimSpace(info), "\n")
	srvChannel <- S_Service{
		Name:        name,
		Description: infoLines[1][3:],
		Running:     strings.Fields(infoLines[0])[2] == "started",
		Enabled:     slices.Contains(enabledList, name),
	}
}

func (mgr *S_OpenRCManager) StartService(name string) error {
	_, err := H_Execute(false, "rc-service", name, "start")
	return err
}

func (mgr *S_OpenRCManager) StopService(name string) error {
	_, err := H_Execute(false, "rc-service", name, "stop")
	return err
}

func (mgr *S_OpenRCManager) RestartService(name string) error {
	_, err := H_Execute(false, "rc-service", name, "restart")
	return err
}

func (mgr *S_OpenRCManager) EnableService(name string) error {
	_, err := H_Execute(false, "rc-update", "add", name, "default")
	return err
}

func (mgr *S_OpenRCManager) DisableService(name string) error {
	_, err := H_Execute(false, "rc-update", "del", name, "default")
	return err
}

// *** systemd
type S_SystemdManager struct{}

func (mgr *S_SystemdManager) IsStub() bool {
	return false
}

func (mgr *S_SystemdManager) Version() string {
	v, err := H_Execute(false, "systemctl", "--version")
	if err != nil {
		return "unknown"
	}

	return strings.Split(strings.TrimSpace(v), "\n")[0]
}

func (mgr *S_SystemdManager) ListServices() ([]S_Service, error) {
	// list of service names, descs, and status
	srvListRaw, err := H_Execute(false, "systemctl", "list-units", "--type=service", "--all", "--no-pager", "--no-legend")
	if err != nil {
		return nil, err
	}

	// list of enabled services (and other bs we dont care about)
	enabledRaw, err := H_Execute(false, "systemctl", "list-unit-files", "--type=service", "--all", "--no-pager", "--no-legend", "--state=enabled")
	if err != nil {
		return nil, err
	}

	srvList := strings.Split(strings.TrimSpace(srvListRaw), "\n")

	// just get the names of enabled services
	var enabledList []string
	for _, line := range strings.Split(strings.TrimSpace(enabledRaw), "\n") {
		enabledList = append(enabledList, strings.Fields(line)[0])
	}

	var services []S_Service
	for _, srv := range srvList {
		// skip the unicode dot that shows up sometimes. TODO: handle this correctly, as the dot means we cant actually do anything
		if srv[:3] == "\u25cf" {
			srv = srv[3:]
		}

		// create S_Service
		fields := strings.Fields(srv)
		services = append(services, S_Service{
			Name:        fields[0],
			Description: strings.Join(fields[4:], " "),
			Running:     fields[3] == "running",
			Enabled:     slices.Contains(enabledList, fields[0]),
		})
	}

	return services, nil
}

func (mgr *S_SystemdManager) StartService(name string) error {
	_, err := H_Execute(false, "systemctl", "start", name)
	return err
}

func (mgr *S_SystemdManager) StopService(name string) error {
	_, err := H_Execute(false, "systemctl", "stop", name)
	return err
}

func (mgr *S_SystemdManager) RestartService(name string) error {
	_, err := H_Execute(false, "systemctl", "restart", name)
	return err
}

func (mgr *S_SystemdManager) EnableService(name string) error {
	_, err := H_Execute(false, "systemctl", "enable", name)
	return err
}

func (mgr *S_SystemdManager) DisableService(name string) error {
	_, err := H_Execute(false, "systemctl", "disable", name)
	return err
}
