//go:build windows

package main

import (
	"context"
	"strings"
	"time"

	"github.com/IceWhaleTech/CasaOS-Common/utils/vfs"
	"github.com/IceWhaleTech/CasaOS-LocalStorage/service"
)

// monitorUEvent watches for drives that are plugged or unplugged. Windows has no
// uevents, so the list of drives is compared periodically.
func monitorUEvent(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	last := strings.Join(vfs.Drives(), "")

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			current := strings.Join(vfs.Drives(), "")
			if current != last {
				last = current
				service.MyService.Disk().RemoveLSBLKCache()
				sendUSBBySocket()
			}
		}
	}
}
