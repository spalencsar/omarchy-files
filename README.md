# Omarchy Files

<div align="center">
  <img src="assets/io.github.spalencsar.OmarchyFiles.svg" alt="Omarchy Files" width="160">

### A file manager for Omarchy

Built with [Quickshell](https://quickshell.org/) and [Go](https://go.dev/)

</div>

Omarchy Files is a standalone Wayland file manager built specifically for
[Omarchy](https://omarchy.org/). A small Go daemon performs file operations and
a Quickshell UI provides a compact, theme-aware desktop experience.

It is an independent fork of [Dux Files](https://github.com/duxlabs-dev/dux-files).
The fork keeps the upstream MIT license and attribution while maintaining its
own Omarchy-focused branding, packaging, and interface.

## Features

**Tabs, panes and views**
Tabs, a split view with two panes, and list, grid and column views. Sort and view can be remembered per folder.

**Fast navigation**
Type to jump to a name or filter the folder, jump to any folder you have visited before, and reopen closed tabs.

**Search**
Recursive search by name with filters. Uses [danksearch](https://github.com/duxlabs-dev/dux-search) when it is running and walks the tree when it isn't.

**File operations**
Copy, move, rename, trash and delete run as jobs with progress and conflict handling, and most of them can be undone. Denied operations offer Retry as administrator through polkit.

**Archives**
Opening an archive extracts it next to itself, like Nautilus. Compress to zip, tar.gz, tar.xz and tar.zst. All pure Go, no external tools.

**Devices and network**
Mount, unmount and eject drives through udisks2. Connect to sftp, smb, ftp, WebDAV, nfs and afp shares through gvfs.

**Trash, recent files and bookmarks**
Browse and restore the trash, see recently used files, and keep bookmarks in the sidebar.

**Open and save dialogs**
Can serve the desktop portal's file chooser, so other apps get the same window, sidebar and keys when they open or save a file.

**Desktop integration**
Owns `org.freedesktop.FileManager1`, so "show in folder" from browsers and other apps lands here.

## Installation on Omarchy

Omarchy Files is supported on Omarchy/Arch Linux. It is distributed as a local
Arch package so the binary, desktop entry, icon, D-Bus files, and user service
are installed and removed together.

### Build and install the Arch package

```bash
git clone --recurse-submodules https://github.com/spalencsar/omarchy-files.git
cd omarchy-files
cd distro/arch
makepkg -Cfs
sudo pacman -U ./omarchy-files-git-*.pkg.tar.zst
```

When pacman offers to reinstall the same package version, confirm it. Development
packages currently use a stable local version number, so this is expected.

Start it once for the current session, or enable the user service for instant
folder opening after login:

```bash
systemctl --user enable --now omarchy-files.service
```

After a package upgrade, restart the existing daemon:

```bash
omarchy-files kill
systemctl --user restart omarchy-files.service
```

### Development build

For development, install the build dependencies available on Omarchy and run:

```bash
make
make test
make test-qml
```

### Optional dependencies

These Arch packages are detected when the daemon starts:

| package | enables |
| --- | --- |
| `bubblewrap` | runs external thumbnailers in a sandbox, like gnome-desktop does |
| `ffmpegthumbnailer` | video thumbnails |
| `poppler` | PDF thumbnails |
| `gvfs` | network shares and the trash on removable drives |
| `udisks2` | the devices sidebar |

Images in png, jpeg, gif, bmp, tiff and webp are decoded in-process. Anything else uses the freedesktop thumbnailers installed on the system, so packages that ship one (glycin, gsf, evince, totem) just work. `omarchy-files ipc system.capabilities` shows what was found.

PDF previews use the optional `QtQuick.Pdf` QML module, supplied by
`qt6-webengine` on Arch. Without it, the preview shows an installation hint;
image previews and the rest of the file manager still work.

## Usage

```bash
omarchy-files show                     # Show the window, starting omarchy-files if needed
omarchy-files open ~/Downloads         # Open folders, one tab each
omarchy-files open --new-window ~/src  # Same, in a new window
omarchy-files select ~/notes.txt       # Open the folder with the file selected
omarchy-files properties ~/notes.txt   # Properties dialog
```

`omarchy-files open` accepts plain paths and these locations:

| location | opens |
| --- | --- |
| `/path`, `relative/path`, `file:///path` | that path |
| `trash:///` | the trash |
| `computer:///` | `/` |
| `network:///` | the Connect to Server dialog |
| `admin:///path` | the path, with Retry as administrator on denied operations |
| `sftp:// ssh:// smb:// ftp:// ftps:// dav:// davs:// nfs:// afp://` | mounted through gvfs, then opened in a new tab |

Started by the systemd unit, closing the window just hides it. Started any other way, closing the last window quits unless Settings, Close behavior says otherwise.

### Command line

The CLI talks to the running omarchy-files when there is one and does the work itself when there isn't. `--json` gives machine-readable output.

```bash
omarchy-files trash put|list|count|restore|empty
omarchy-files mime query|default|set
omarchy-files devices list|mount|unmount|eject
omarchy-files search <root> <query>
omarchy-files ipc list                 # Every IPC method the daemon exposes
omarchy-files ipc <method> key=value...
```

Jobs started from the CLI answer their own prompts. A restore that finds its original path taken keeps both (`--conflict skip|replace` changes that), and a failed item is skipped, reported, and makes the command exit 1.

## Default file manager

To open folders here from other apps:

```bash
xdg-mime default io.github.spalencsar.OmarchyFiles.desktop inode/directory
```

The installer doesn't do this for you.

While it runs, omarchy-files owns `org.freedesktop.FileManager1`, so `ShowFolders`, `ShowItems` and `ShowItemProperties` calls come here. Chromium based browsers go through the portal's `OpenDirectory`, which forwards to `ShowItems`. omarchy-files never takes the name from another file manager, it waits in line until that one quits. Turn it off with Settings, General, "Register as the session file manager".

`make install` also ships an activation file for the name, so a call can start omarchy-files when nothing owns it. Nautilus, Dolphin, Thunar and Nemo ship one too. dbus-daemon prefers the first service directory, and the default `PREFIX=/usr/local` comes before `/usr/share`.

## Open and save dialogs

Nothing changes until you turn it on. Use Settings, General, "Use Omarchy Files for open and save dialogs", or:

```bash
omarchy-files portal install
omarchy-files portal status
omarchy-files portal uninstall   # put things back
```

This puts omarchy-files first for the FileChooser in your `~/.config/xdg-desktop-portal/<desktop>-portals.conf`, keeps the old backend as a fallback, and remembers what it changed. xdg-desktop-portal only reads its config at startup, so it has to restart. Both the settings page and the CLI ask first, since a restart cancels dialogs other apps have open.

Portal-aware applications pick it up, including Chromium browsers, Firefox with
`widget.use-xdg-desktop-portal.file-picker` set to `1`, GTK 4 applications with
`GDK_DEBUG=portals`, and Qt applications with
`QT_QPA_PLATFORMTHEME=xdgdesktopportal`. Applications that draw their own
dialogs are unchanged.

The dialog is its own window. Qt 6.12 and Quickshell cannot attach it to the
calling app's window, so `modal` is ignored. Without the user service the first
dialog is a cold start.

The Omarchy package installs the portal metadata and activation service. It does
not select Omarchy Files as the portal backend until you explicitly enable it.

## Repository Structure

```
omarchy-files/
├── core/               # Go daemon and the omarchy-files CLI
│   ├── cmd/omarchy-files/     # CLI commands and the daemon entrypoint
│   └── internal/       # File ops, IPC, devices, search, portal, D-Bus
├── quickshell/         # The UI
│   ├── Common/         # Theme, settings, i18n
│   ├── Services/       # Bridges to the daemon
│   ├── Modules/        # Windows, panes, dialogs, settings
│   └── translations/   # POEditor catalogs
├── dux-qml-common/    # Shared widget library (submodule, linked as quickshell/DCommon)
├── distro/             # Arch packaging and development distribution files
└── assets/             # Icon, desktop entry, D-Bus, portal and systemd files
```

## Development

```bash
make run                                                  # Build and run against the repo's QML
OMARCHY_FILES_ENABLE_HOTRELOAD=1 core/bin/omarchy-files run -c quickshell  # Hot reload the UI
make test                                                 # Go tests
make test-qml                                             # QML tests
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for setup, the shared widget library,
translations, and contributor guidance.

## License

MIT License - See [LICENSE](LICENSE) for details.
