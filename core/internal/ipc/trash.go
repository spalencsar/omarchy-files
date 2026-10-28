package ipc

import (
	"cmp"
	"context"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"time"

	"github.com/duxlabs-dev/dux-go/files"
	"github.com/duxlabs-dev/dux-go/paths"
	"github.com/duxlabs-dev/dux-go/trash"
	"github.com/spalencsar/omarchy-files/core/internal/fileops"
)

const trashDateLayout = "2006-01-02T15:04:05"

type TrashItem struct {
	files.Entry
	TrashName    string `json:"trashName"`
	TrashDir     string `json:"trashDir"`
	OriginalPath string `json:"originalPath"`
	DeletedMs    int64  `json:"deletedMs"`
}

var listTrash = trash.List

func HandleTrash(_ context.Context, w *ConnWriter, req Request, deps Deps) {
	switch req.Method {
	case "trash.list":
		entries, err := listTrash()
		if err != nil {
			w.RespondError(req.ID, "trash.list: "+err.Error())
			return
		}
		if entries == nil {
			entries = []trash.Entry{}
		}
		w.Respond(req.ID, map[string]any{"entries": entries})
	case "trash.entries":
		if deps.Files == nil {
			w.RespondError(req.ID, "file service unavailable")
			return
		}
		entries, err := listTrash()
		if err != nil {
			w.RespondError(req.ID, "trash.entries: "+err.Error())
			return
		}
		w.Respond(req.ID, map[string]any{"entries": trashItems(entries, deps.Files.Stat)})
	case "trash.count":
		count, err := trash.Count()
		if err != nil {
			w.RespondError(req.ID, "trash.count: "+err.Error())
			return
		}
		w.Respond(req.ID, map[string]any{"count": count, "stamp": trashStamp(count)})
	default:
		w.RespondError(req.ID, "unknown trash method: "+req.Method)
	}
}

func trashItems(entries []trash.Entry, stat func(string) (files.Entry, error)) []TrashItem {
	items := make([]TrashItem, 0, len(entries))
	for _, e := range entries {
		entry, err := stat(e.FilesPath)
		if err != nil {
			continue
		}
		original, _ := fileops.RestorePath(e.TrashDir, e.OriginalPath)
		if original != "" {
			entry.Name = filepath.Base(original)
		}
		items = append(items, TrashItem{
			Entry:        entry,
			TrashName:    e.Name,
			TrashDir:     e.TrashDir,
			OriginalPath: original,
			DeletedMs:    deletedMs(e),
		})
	}
	slices.SortStableFunc(items, func(a, b TrashItem) int {
		return cmp.Or(cmp.Compare(b.DeletedMs, a.DeletedMs), cmp.Compare(a.Name, b.Name))
	})
	return items
}

func trashStamp(count int) string {
	stamp := strconv.Itoa(count)
	info, err := os.Stat(filepath.Join(paths.XDGDataHome(), "Trash", "info"))
	if err != nil {
		return stamp
	}
	return stamp + ":" + strconv.FormatInt(info.ModTime().UnixNano(), 10)
}

func deletedMs(e trash.Entry) int64 {
	if when, err := time.ParseInLocation(trashDateLayout, e.DeletionDate, time.Local); err == nil {
		return when.UnixMilli()
	}
	info, err := os.Lstat(e.InfoPath)
	if err != nil {
		return 0
	}
	return info.ModTime().UnixMilli()
}
