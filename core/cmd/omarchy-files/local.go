package main

import (
	"os"

	"github.com/spalencsar/omarchy-files/core/internal/devices"
	"github.com/spalencsar/omarchy-files/core/internal/fileops"
	"github.com/spalencsar/omarchy-files/core/internal/ipc"
	"github.com/spalencsar/omarchy-files/core/internal/mimeapps"
	"github.com/spalencsar/omarchy-files/core/internal/sandbox"
	"github.com/spalencsar/omarchy-files/core/internal/search"
	"github.com/spalencsar/omarchy-files/core/internal/settings"
	"github.com/duxlabs-dev/dux-go/files"
	"github.com/duxlabs-dev/dux-go/log"
)

func noServices(*ipc.EventBus) (ipc.Deps, func()) { return ipc.Deps{}, func() {} }

func appServices(*ipc.EventBus) (ipc.Deps, func()) {
	return ipc.Deps{Apps: mimeapps.NewService()}, func() {}
}

func fileServices(bus *ipc.EventBus) (ipc.Deps, func()) {
	apps := mimeapps.NewService()
	fileService := files.NewService(bus, cacheHome(), apps.Detector())
	return ipc.Deps{Apps: apps, Files: fileService, FileOps: fileops.NewManager(bus, fileService)}, fileService.Close
}

func searchServices(bus *ipc.EventBus) (ipc.Deps, func()) {
	fileService := files.NewService(bus, cacheHome(), mimeapps.NewService().Detector())
	deviceService := devices.NewService(bus, uint32(os.Getuid()))
	if err := deviceService.Start(); err != nil {
		log.Debugf("udisks2 unavailable: %v", err)
	}
	searchService := search.NewService(bus, fileService)
	searchService.SetIndexEnabled(!sandbox.Current().Sandboxed)
	searchService.SetMounts(func() []string { return localMounts(deviceService.Snapshot()) })
	searchService.SetExcludes(func() []string { return settings.Load(configHome()).SearchExcludes })
	return ipc.Deps{Files: fileService, Search: searchService}, func() {
		searchService.Close()
		deviceService.Close()
		fileService.Close()
	}
}

func deviceServices(bus *ipc.EventBus) (ipc.Deps, func()) {
	deviceService := devices.NewService(bus, uint32(os.Getuid()))
	if err := deviceService.Start(); err != nil {
		log.Debugf("udisks2 unavailable: %v", err)
	}
	return ipc.Deps{Devices: deviceService}, deviceService.Close
}
