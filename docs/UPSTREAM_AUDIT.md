# Upstream integration audit

Audited against fork commit `46c20fe5` on 2026-10-23 after a clean local build,
Go test run, and QML test run on Omarchy.

## Baseline

The upstream project builds a Go daemon/CLI with an embedded Quickshell UI. The
daemon, QML UI, portal backend, FileManager1 registration, and polkit helper
currently share the `dfiles` identity. `dux-qml-common` is included as a
submodule at `802b4b1`.

The project has a source requirement of Go 1.27 and must build with
`CGO_ENABLED=0`. Its existing `make`, `make test`, and `make test-qml` commands
pass locally before Omarchy-specific changes.

## External identities to migrate together

| Surface | Current value |
| --- | --- |
| CLI/binary and user service | `dfiles` |
| Application ID, desktop file, icon, metainfo, polkit policy | `dev.duxlabs.files` |
| File-manager activation service | `dev.duxlabs.files.FileManager1.service` |
| Shared FileManager1 bus name | `org.freedesktop.FileManager1` |
| Portal backend/service | `dfiles` / `org.freedesktop.impl.portal.desktop.dfiles` |
| Go module/import path | `github.com/duxlabs-dev/dux-files/core` |
| Configuration and portal state | `dfiles` |
| Display name and attribution | Dux Files / Dux Labs |

The shared `org.freedesktop.FileManager1` bus name deliberately remains
unchanged: it is the desktop standard that any chosen default file manager
claims while active. It must be tested for coexistence, not renamed.

## Recommended project-owned identity

Until written authorization to use an official Omarchy namespace exists, use a
project-owned identity that matches this fork:

| Surface | Recommended value |
| --- | --- |
| Display name | Omarchy Files |
| CLI/binary and user service | `omarchy-files` |
| Application ID / desktop ID / icon | `io.github.spalencsar.OmarchyFiles` |
| File-manager activation file | `io.github.spalencsar.OmarchyFiles.FileManager1.service` |
| Portal backend | `omarchy-files` |
| Portal D-Bus service | `org.freedesktop.impl.portal.desktop.omarchy-files` |
| Go module | `github.com/spalencsar/omarchy-files/core` |
| User configuration/state | `omarchy-files` |

Do not use `org.omarchy.Files` unless its namespace owner explicitly approves
it. This prevents the fork from implying that it is shipped or endorsed by
Omarchy.

## Packaging and desktop integration

The root `Makefile` currently installs a binary, desktop entry, icon, AppStream
metadata, polkit policy, FileManager1 D-Bus activation, portal files, and an
optional systemd user service. The supported Omarchy package must stage all of
these via a package build; `sudo make install` is not a release path.

The portal backend is packaged but not selected until an explicit user action.
That behavior should remain unchanged. Changing portal preference or restarting
`xdg-desktop-portal` remains an opt-in milestone after identity and theme work.

## First change set

1. Rename every external identity above atomically, including generated paths,
   tests, source imports, release automation, Nix and Flatpak manifests.
2. Preserve the upstream MIT license and attribution; update application
   metadata to identify this fork accurately.
3. Run `make`, `make test`, `make test-qml`, plus `git diff --check`.
4. Confirm the upstream and fork can be installed side by side before default
   MIME or portal behavior is touched.

No visual theme work, Hyprland configuration, MIME default, D-Bus ownership,
portal preference, or system-service activation is part of this first change
set.
