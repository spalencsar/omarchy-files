package ipc

import (
	"context"
	"os/exec"

	duxipc "github.com/duxlabs-dev/dux-go/ipc"
	"github.com/duxlabs-dev/dux-go/ipc/params"
	"github.com/spalencsar/omarchy-files/core/internal/filemanager1"
	"github.com/spalencsar/omarchy-files/core/internal/portal"
	"github.com/spalencsar/omarchy-files/core/internal/sandbox"
)

func HandleSystem(ctx context.Context, w *ConnWriter, req Request, deps Deps) {
	switch req.Method {
	case "system.colorScheme.get":
		if deps.ColorScheme == nil {
			w.Respond(req.ID, map[string]any{"available": false, "colorScheme": uint32(0)})
			return
		}
		w.Respond(req.ID, deps.ColorScheme.State())
	case "system.capabilities":
		w.Respond(req.ID, capabilities(deps))
	case "system.elevation":
		if deps.FileOps == nil {
			w.Respond(req.ID, map[string]any{"available": false, "reason": "NOROUTE"})
			return
		}
		w.Respond(req.ID, deps.FileOps.Elevation())
	case "system.fileManager.get":
		if deps.FileManager == nil {
			w.Respond(req.ID, filemanager1.Status{Reason: "only the window process registers"})
			return
		}
		w.Respond(req.ID, deps.FileManager.Status())
	case "system.fileManager.set":
		if deps.FileManager == nil {
			w.RespondError(req.ID, "only the window process registers as the file manager")
			return
		}
		w.Respond(req.ID, deps.FileManager.SetEnabled(params.BoolLoose(req.Params, "enabled")))
	case "system.portal.get":
		if deps.Portal == nil {
			w.Respond(req.ID, map[string]any{"available": false, "reason": portal.ReasonNoWindow})
			return
		}
		w.Respond(req.ID, deps.Portal.Status())
	case "system.portal.install", "system.portal.uninstall":
		if deps.Portal == nil {
			w.RespondError(req.ID, "only the window process serves the portal")
			return
		}
		change := deps.Portal.Install
		if req.Method == "system.portal.uninstall" {
			change = deps.Portal.Uninstall
		}
		report, err := change()
		if err != nil {
			w.RespondError(req.ID, err.Error())
			return
		}
		w.Respond(req.ID, report)
	case "system.portal.restart":
		if deps.Portal == nil {
			w.RespondError(req.ID, "only the window process serves the portal")
			return
		}
		if err := deps.Portal.Restart(); err != nil {
			w.RespondError(req.ID, err.Error())
			return
		}
		w.Respond(req.ID, map[string]any{"restarted": true})
	case "system.openUri":
		if deps.Apps == nil {
			w.RespondError(req.ID, "uri opener unavailable")
			return
		}
		uri := params.StringOpt(req.Params, "uri", "")
		if uri == "" {
			w.RespondError(req.ID, "uri is required")
			return
		}
		if err := deps.Apps.OpenURI(ctx, uri); err != nil {
			w.RespondError(req.ID, err.Error())
			return
		}
		w.Respond(req.ID, map[string]any{"opened": true})
	default:
		w.RespondError(req.ID, "unknown system method: "+req.Method)
	}
}

type capabilitiesResponse struct {
	sandbox.Capabilities
	Thumbnailers         []string                      `json:"thumbnailers"`
	OptionalDependencies map[string]optionalDependency `json:"optionalDependencies"`
}

type optionalDependency struct {
	Installed bool `json:"installed"`
	Active    bool `json:"active,omitempty"`
}

func capabilities(deps Deps) capabilitiesResponse {
	caps := deps.Sandbox
	if deps.FileOps != nil {
		e := deps.FileOps.Elevation()
		caps = caps.WithFeature("elevation", sandbox.Feature{Available: e.Available, Reason: e.Reason})
	}
	if caps.Portals == nil {
		caps.Portals = map[string]uint32{}
	}
	if caps.Features == nil {
		caps.Features = map[string]sandbox.Feature{}
	}
	if caps.Filesystems == nil {
		caps.Filesystems = []string{}
	}
	out := capabilitiesResponse{
		Capabilities: caps,
		Thumbnailers: []string{},
		OptionalDependencies: map[string]optionalDependency{
			"dsearch":           dsearchDependency(),
			"bubblewrap":        {Installed: commandAvailable("bwrap")},
			"ffmpegthumbnailer": {Installed: commandAvailable("ffmpegthumbnailer")},
			"poppler":           {Installed: commandAvailable("pdftoppm")},
		},
	}
	if deps.Files != nil {
		out.Thumbnailers = append(out.Thumbnailers, deps.Files.Capabilities()...)
	}
	return out
}

func commandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func dsearchDependency() optionalDependency {
	dependency := optionalDependency{Installed: commandAvailable("dsearch")}
	if !dependency.Installed {
		return dependency
	}
	_, err := duxipc.FindRunningSocket("danksearch")
	dependency.Active = err == nil
	return dependency
}
