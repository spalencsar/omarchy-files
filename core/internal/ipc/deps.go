package ipc

import (
	"github.com/duxlabs-dev/dux-go/files"
	"github.com/spalencsar/omarchy-files/core/internal/bookmarks"
	"github.com/spalencsar/omarchy-files/core/internal/clipboard"
	"github.com/spalencsar/omarchy-files/core/internal/colorscheme"
	"github.com/spalencsar/omarchy-files/core/internal/devices"
	"github.com/spalencsar/omarchy-files/core/internal/filemanager1"
	"github.com/spalencsar/omarchy-files/core/internal/fileops"
	"github.com/spalencsar/omarchy-files/core/internal/gitstatus"
	"github.com/spalencsar/omarchy-files/core/internal/mimeapps"
	"github.com/spalencsar/omarchy-files/core/internal/netmount"
	"github.com/spalencsar/omarchy-files/core/internal/portal"
	"github.com/spalencsar/omarchy-files/core/internal/recent"
	"github.com/spalencsar/omarchy-files/core/internal/sandbox"
	"github.com/spalencsar/omarchy-files/core/internal/search"
	"github.com/spalencsar/omarchy-files/core/internal/settings"
)

type Deps struct {
	Bus         *EventBus
	Pending     *PendingOpen
	Version     string
	ColorScheme *colorscheme.Watcher
	Files       *files.Service
	FileOps     *fileops.Manager
	Git         *gitstatus.Service
	Apps        *mimeapps.Service
	Devices     *devices.Service
	Network     *netmount.Service
	Clipboard   *clipboard.Service
	Search      *search.Service
	Bookmarks   *bookmarks.Store
	Recent      *recent.Store
	Settings    func() settings.UISettings
	Sandbox     sandbox.Capabilities
	FileManager *filemanager1.Service
	Portal      *portal.Controller
	Quit        func()
}
