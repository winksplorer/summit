package main

import (
	"log"

	"github.com/anatol/smart.go"
	"github.com/jaypipes/ghw"
)

type ST_Disk struct {
	Name       string `msgpack:"name"`       // name (nvme0n1, loop0, etc.)
	Size       uint64 `msgpack:"size"`       // size
	Type       string `msgpack:"type"`       // hdd, fdd, odd, or ssd
	Controller string `msgpack:"controller"` // scsi, ide, virtio, mmc, or nvme
	Removable  bool   `msgpack:"removable"`  // removable?
	Vendor     string `msgpack:"vendor"`     // vendor string
	Model      string `msgpack:"model"`      // model string
	Serial     string `msgpack:"serial"`     // serial number

	Partitions []ST_Partition `msgpack:"partitions"` // partitions
	SMART      ST_SMART       `msgpack:"smart"`      // SMART data
}

type ST_Partition struct {
	Name       string `msgpack:"name"`       // name (nvme0n1p1, sda1, etc.)
	FsLabel    string `msgpack:"fST_label"`  // filesystem label
	Size       uint64 `msgpack:"size"`       // size
	Type       string `msgpack:"type"`       // filesystem type
	Mountpoint string `msgpack:"mountpoint"` // mount point
	Readonly   bool   `msgpack:"ro"`         // readonly?
	UUID       string `msgpack:"uuid"`       // part uuid
}

type ST_SMART struct {
	// generic
	DataAvailable bool   `msgpack:"available"`      // Is SMART data even available for this disk?
	Temperature   uint64 `msgpack:"temperature"`    // SMART: Temperature in Celsius
	Read          uint64 `msgpack:"read"`           // SMART: Data units (LBA) read
	Written       uint64 `msgpack:"written"`        // SMART: Data units (LBA) written
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

func ST_AssembleSMART(diskPath string) (ST_SMART, error) {
	var SMART ST_SMART

	a := &smart.GenericAttributes{}
	dev, err := smart.Open(diskPath)
	if err != nil {
		return ST_SMART{}, err
	}

	// generic attrs
	if a, err = dev.ReadGenericAttributes(); err != nil {
		return ST_SMART{}, err
	}

	// base struct
	SMART = ST_SMART{
		DataAvailable: true,
		Temperature:   a.Temperature,
		Read:          a.Read,
		Written:       a.Written,
		PowerOnHours:  a.PowerOnHours,
		PowerCycles:   a.PowerCycles,
	}

	// controller-specific attrs
	switch sm := dev.(type) {
	case *smart.SataDevice: // ATA
		data, err := sm.ReadSMARTData()
		if err != nil {
			return ST_SMART{}, err
		}

		for _, attr := range data.Attrs {
			switch attr.Name {
			case "Reallocate_NAND_Blk_Cnt", "Reallocated_Sector_Ct":
				SMART.AtaReallocSectors = attr.ValueRaw
			case "Offline_Uncorrectable", "Reported_Uncorrectable_Errors", "Uncorrectable_Error_Cnt":
				SMART.AtaUncorrectableErrs = attr.ValueRaw
			}
		}
	case *smart.NVMeDevice: // NVMe
		data, err := sm.ReadSMART()
		if err != nil {
			return ST_SMART{}, err
		}

		SMART.NvmeCritWarning = data.CritWarning
		SMART.NvmeAvailSpare = data.AvailSpare
		SMART.NvmePercentUsed = data.PercentUsed
		SMART.NvmeUnsafeShutdowns = data.UnsafeShutdowns.Val[0]
		SMART.NvmeMediaErrs = data.MediaErrors.Val[0]
	}

	return SMART, nil
}

func ST_GetDevices() ([]ST_Disk, error) {
	var disks []ST_Disk

	blocks, err := ghw.Block()
	if err != nil {
		return nil, err
	}

	for _, disk := range blocks.Disks {
		var parts []ST_Partition

		SMART, err := ST_AssembleSMART("/dev/" + disk.Name)
		if err != nil {
			log.Println("ST_AssembleSMART: /dev/"+disk.Name+":", err)
		}

		for _, part := range disk.Partitions {
			parts = append(parts, ST_Partition{
				Name:       part.Name,
				FsLabel:    part.FilesystemLabel,
				Size:       part.SizeBytes,
				Type:       part.Type,
				Mountpoint: part.MountPoint,
				Readonly:   part.IsReadOnly,
				UUID:       part.UUID,
			})
		}

		disks = append(disks, ST_Disk{
			Name:       disk.Name,
			Size:       disk.SizeBytes,
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
	return ST_GetDevices()
}
