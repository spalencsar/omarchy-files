package main

import (
	"time"

	"github.com/duxlabs-dev/dux-go/log"
	dankpaths "github.com/duxlabs-dev/dux-go/paths"
	"github.com/spalencsar/omarchy-files/core/internal/filechooser"
	"github.com/spalencsar/omarchy-files/core/internal/mimeapps"
	"github.com/spalencsar/omarchy-files/core/internal/recent"
	"github.com/spalencsar/omarchy-files/core/internal/settings"
)

const (
	recentExec        = "'gio open %u'"
	recentFallbackApp = "omarchy-files"
	recentUnknownMime = "application/octet-stream"
)

func wireRecent(apps *mimeapps.Service, chooser *filechooser.Service, config string) *recent.Store {
	store := recent.NewStore(dankpaths.XDGDataHome())
	record := func(paths, mimes []string, app string) {
		if len(paths) == 0 || !settings.Load(config).RecentFiles {
			return
		}
		uses := make([]recent.Use, len(paths))
		for i, path := range paths {
			mimeType := ""
			if i < len(mimes) {
				mimeType = mimes[i]
			}
			if mimeType == "" {
				mimeType, _ = apps.MimeOf(path)
			}
			if mimeType == "" {
				mimeType = recentUnknownMime
			}
			uses[i] = recent.Use{Path: path, MimeType: mimeType}
		}
		if err := store.Record(uses, app, recentExec, time.Now()); err != nil {
			log.Warnf("recent list not updated: %v", err)
		}
	}
	apps.Opened = func(paths, mimes []string, app *mimeapps.Entry) {
		name := recentFallbackApp
		if app != nil && app.Name != "" {
			name = app.Name
		}
		record(paths, mimes, name)
	}
	if chooser != nil {
		chooser.Chosen = func(req filechooser.Request, paths []string) {
			name := req.AppID
			if name == "" {
				name = recentFallbackApp
			}
			record(paths, nil, name)
		}
	}
	return store
}
