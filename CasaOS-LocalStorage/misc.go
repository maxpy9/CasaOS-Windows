package main

import (
	"reflect"
	"strconv"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-LocalStorage/model"
	"github.com/IceWhaleTech/CasaOS-LocalStorage/service"
	"go.uber.org/zap"
)

func sendDiskBySocket() {
	blkList := service.MyService.Disk().LSBLK(true)

	status := model.DiskStatus{}
	healthy := true

	//var systemDisk *model.LSBLKModel

	for _, currentDisk := range blkList {

		// if systemDisk == nil {
		// 	// go 5 level deep to look for system block device by mount point being "/"
		// 	systemDisk = service.WalkDisk(currentDisk, 5, func(blk model.LSBLKModel) bool { return blk.MountPoint == "/" })

		// 	if systemDisk != nil {
		// 		s, _ := strconv.ParseUint(systemDisk.FSSize.String(), 10, 64)
		// 		a, _ := strconv.ParseUint(systemDisk.FSAvail.String(), 10, 64)
		// 		u, _ := strconv.ParseUint(systemDisk.FSUsed.String(), 10, 64)
		// 		status.Size += s
		// 		status.Avail += a
		// 		status.Used += u
		//		continue
		// 	}
		// }
		if !service.IsDiskSupported(currentDisk) {
			continue
		}
		temp := service.MyService.Disk().SmartCTL(currentDisk.Path)
		if reflect.DeepEqual(temp, model.SmartctlA{}) {
			healthy = true
		} else {
			if len(temp.ModelName) > 0 {
				healthy = temp.SmartStatus.Passed
			} else {
				healthy = true
			}
		}
		if len(currentDisk.Children) > 0 {
			for _, v := range currentDisk.Children {
				if len(v.MountPoint) > 0 {
					s, _ := strconv.ParseUint(v.FSSize.String(), 10, 64)
					a, _ := strconv.ParseUint(v.FSAvail.String(), 10, 64)
					u, _ := strconv.ParseUint(v.FSUsed.String(), 10, 64)
					status.Size += s
					status.Avail += a
					status.Used += u
				}
			}
		} else {
			if len(currentDisk.MountPoint) > 0 {
				s, _ := strconv.ParseUint(currentDisk.FSSize.String(), 10, 64)
				a, _ := strconv.ParseUint(currentDisk.FSAvail.String(), 10, 64)
				u, _ := strconv.ParseUint(currentDisk.FSUsed.String(), 10, 64)
				status.Size += s
				status.Avail += a
				status.Used += u
			}
		}
	}

	status.Health = healthy
	message := make(map[string]interface{})
	message["sys_disk"] = status
	if err := service.MyService.NotifySystem().SendSystemStatusNotify(message); err != nil {
		logger.Error("failed to send notify", zap.Any("message", message), zap.Error(err))
	}
}

func sendUSBBySocket() {
	message := map[string]interface{}{
		"sys_usb": service.MyService.Disk().GetUSBDriveStatusList(),
	}

	if err := service.MyService.NotifySystem().SendSystemStatusNotify(message); err != nil {
		logger.Error("failed to send notify", zap.Any("message", message), zap.Error(err))
	}
}

func sendStorageStats() {
	sendDiskBySocket()
	sendUSBBySocket()
}
