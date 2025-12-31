package main

import (
	"log"

	"github.com/anatol/smart.go"
	"github.com/jaypipes/ghw"
)

type S_Disk struct {
	Name       string `msgpack:"name"`       // name (nvme0n1, loop0, etc.)
	Size       string `msgpack:"size"`       // size, human readable
	Type       string `msgpack:"type"`       // hdd, fdd, odd, or ssd
	Controller string `msgpack:"controller"` // scsi, ide, virtio, mmc, or nvme
	Removable  bool   `msgpack:"removable"`  // removable?
	Vendor     string `msgpack:"vendor"`     // vendor string
	Model      string `msgpack:"model"`      // model string
	Serial     string `msgpack:"serial"`     // serial number

	Partitions []S_Partition `msgpack:"partitions"` // partitions
	SMART      S_SMART       `msgpack:"smart"`      // SMART data
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

type S_SMART struct {
	// generic
	DataAvailable bool   `msgpack:"available"`      // Is SMART data even available for this disk?
	Temperature   uint64 `msgpack:"temperature"`    // SMART: Temperature in Celsius
	Read          string `msgpack:"read"`           // SMART: Data units (LBA) read, human readable
	Written       string `msgpack:"written"`        // SMART: Data units (LBA) written, human readable
	PowerOnHours  uint64 `msgpack:"power_on_hours"` // SMART: Power on time in hours
	PowerCycles   uint64 `msgpack:"power_cycles"`   // SMART: Power cycles

	// ata
	AtaReallocSectors    uint64 `msgpack:"ata_realloc_sectors"`    // SMART/ATA: Reallocated_Sector_Ct
	AtaUncorrectableErrs uint64 `msgpack:"ata_uncorrectable_errs"` // SMART/ATA: Uncorrectable_Error_Cnt

	// nvme
	NvmeCritWarning     uint8  `msgpack:"nvme_crit_warning"`     // SMART/NVMe: Critical Warning
	NvmeAvailSpare      uint8  `msgpack:"nvme_avail_spare"`      // SMART/NVMe: Available Spare
	NvmePercentUsed     uint8  `msgpack:"nvme_percent_used"`     // SMART/NVMe: Percentage Used
	NvmeUnsafeShutdowns uint64 `msgpack:"nvme_unsafe_shutdowns"` // SMART/NVMe: Unexpected Power Losses
	NvmeMediaErrs       uint64 `msgpack:"nvme_media_errs"`       // SMART/NVMe: Media and Data Integrity Errors
}

func S_AssembleSMART(diskPath string) (S_SMART, error) {
	var SMART S_SMART

	a := &smart.GenericAttributes{}
	dev, err := smart.Open(diskPath)
	if err != nil {
		return S_SMART{}, err
	} else {
		// generic attrs
		if a, err = dev.ReadGenericAttributes(); err != nil {
			return S_SMART{}, err
		}

		// base struct
		SMART = S_SMART{
			DataAvailable: true,
			Temperature:   a.Temperature,
			Read:          H_HumanReadable(a.Read),
			Written:       H_HumanReadable(a.Written),
			PowerOnHours:  a.PowerOnHours,
			PowerCycles:   a.PowerCycles,
		}

		// controller-specific attrs
		switch sm := dev.(type) {
		case *smart.SataDevice: // ATA
			data, err := sm.ReadSMARTData()
			if err != nil {
				return S_SMART{}, err
			}

			for _, attr := range data.Attrs {
				switch attr.Name {
				case "Reallocate_NAND_Blk_Cnt":
					fallthrough
				case "Reallocated_Sector_Ct":
					SMART.AtaReallocSectors = attr.ValueRaw
				case "Offline_Uncorrectable":
					fallthrough
				case "Reported_Uncorrectable_Errors":
					fallthrough
				case "Uncorrectable_Error_Cnt":
					SMART.AtaUncorrectableErrs = attr.ValueRaw
				}
			}
		case *smart.NVMeDevice: // NVMe
			data, err := sm.ReadSMART()
			if err != nil {
				return S_SMART{}, err
			}

			SMART.NvmeCritWarning = data.CritWarning
			SMART.NvmeAvailSpare = data.AvailSpare
			SMART.NvmePercentUsed = data.PercentUsed
			SMART.NvmeUnsafeShutdowns = data.UnsafeShutdowns.Val[0]
			SMART.NvmeMediaErrs = data.MediaErrors.Val[0]
		}
	}

	return SMART, nil
}

func S_GetDevices() ([]S_Disk, error) {
	var disks []S_Disk

	blocks, err := ghw.Block()
	if err != nil {
		return nil, err
	}

	for _, disk := range blocks.Disks {
		var parts []S_Partition

		SMART, err := S_AssembleSMART("/dev/" + disk.Name)
		if err != nil {
			log.Println("S_AssembleSMART: /dev/"+disk.Name+":", err)
		}

		for _, part := range disk.Partitions {
			parts = append(parts, S_Partition{
				Name:       part.Name,
				FsLabel:    part.FilesystemLabel,
				Size:       H_HumanReadableBytes(part.SizeBytes, 1024),
				Type:       part.Type,
				Mountpoint: part.MountPoint,
				Readonly:   part.IsReadOnly,
				UUID:       part.UUID,
			})
		}

		disks = append(disks, S_Disk{
			Name:       disk.Name,
			Size:       H_HumanReadableBytes(disk.SizeBytes, 1024),
			Type:       disk.DriveType.String(),
			Controller: disk.StorageController.String(),
			Removable:  disk.IsRemovable,
			Vendor:     disk.Vendor,
			Model:      disk.Model,
			Serial:     disk.SerialNumber,
			Partitions: parts,
			SMART:      SMART,
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
