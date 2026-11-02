//go:build !windows

package common

import (
	"time"

	"github.com/IceWhaleTech/CasaOS-LocalStorage/codegen/message_bus"
	"github.com/pilebones/go-udev/netlink"
)

func EventAdapter(e netlink.UEvent) *message_bus.Event {
	devType := e.Env["DEVTYPE"]

	eventType, ok := EventTypes[devType][string(e.Action)]
	if !ok {
		return nil
	}
	time.Sleep(time.Second * 3)
	properties := make(map[string]string)
	for propertyName, envName := range PropertyNameLookupMaps[devType] {
		value, ok := e.Env[envName]
		if !ok {
			continue
		}
		properties[propertyName] = value

	}

	return &message_bus.Event{
		SourceID:   eventType.SourceID,
		Name:       eventType.Name,
		Properties: properties,
	}
}
