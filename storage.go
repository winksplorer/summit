package main

import (
	"github.com/jaypipes/ghw"
)

type S_Disk struct {
	Name       string        `msgpack:"name"`       // name (nvme0n1, loop0, etc.)
	Size       string        `msgpack:"size"`       // size, human readable
	Type       string        `msgpack:"type"`       // hdd, fdd, odd, or ssd
	Controller string        `msgpack:"controller"` // scsi, ide, virtio, mmc, or nvme
	Removable  bool          `msgpack:"removable"`  // removable?
	Vendor     string        `msgpack:"vendor"`     // vendor string
	Model      string        `msgpack:"model"`      // model string
	Serial     string        `msgpack:"serial"`     // serial number
	Partitions []S_Partition `msgpack:"partitions"` // partitions
}

type S_Partition struct {
	Name       string `msgpack:"name"`       // device name (nvme0n1p1, sda1, etc.)
	FsLabel    string `msgpack:"fsLabel"`    // filesystem label
	Size       string `msgpack:"size"`       // device size, human readable
	Type       string `msgpack:"type"`       // filesystem type
	Mountpoint string `msgpack:"mountpoint"` // mount point
	Readonly   bool   `msgpack:"ro"`         // readonly?
	UUID       string `msgpack:"uuid"`       // part uuid
}

func S_GetDevices() ([]S_Disk, error) {
	var disks []S_Disk

	blocks, err := ghw.Block()
	if err != nil {
		return nil, err
	}

	for _, disk := range blocks.Disks {
		var parts []S_Partition

		for _, part := range disk.Partitions {
			parts = append(parts, S_Partition{
				Name:       part.Name,
				FsLabel:    part.FilesystemLabel,
				Size:       H_HumanReadable(part.SizeBytes),
				Type:       part.Type,
				Mountpoint: part.MountPoint,
				Readonly:   part.IsReadOnly,
				UUID:       part.UUID,
			})
		}

		disks = append(disks, S_Disk{
			Name:       disk.Name,
			Size:       H_HumanReadable(disk.SizeBytes),
			Type:       disk.DriveType.String(),
			Controller: disk.StorageController.String(),
			Removable:  disk.IsRemovable,
			Vendor:     disk.Vendor,
			Model:      disk.Model,
			Serial:     disk.SerialNumber,
			Partitions: parts,
		})
	}

	return disks, nil
}

func Comm_StorageGetdevs(data Comm_Message, keyCookie string) (any, error) {
	disks, err := S_GetDevices()
	if err != nil {
		return nil, err
	}

	return disks, nil
}
