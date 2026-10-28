# Changelog

All notable changes to Omarchy Files are documented here.

## [Unreleased]

### Added

- A persistent transfer indicator in the lower-left corner. Its compact
  progress remains visible across folder and tab navigation, while a click
  expands job details and cancellation controls.
- One Omarchy desktop notification when a copy, move, archive, extract,
  duplicate, link, or folder-creation transfer completes, is cancelled, or
  fails. Clicking the notification brings Omarchy Files back into view.

### Fixed

- Protect app-owned configuration, session history, saved network locations,
  bookmarks, and thumbnail caches from other local users. Existing storage is
  migrated to owner-only permissions when the daemon starts.
- Rebase running file operations onto a GVfs share's current mountpoint after
  a remount, and report a disconnected network location explicitly when the
  share is no longer available.
- Report permanent-delete progress entry by entry instead of appearing frozen
  while one monolithic recursive removal blocks on a slow NFS or GVfs mount.

## [0.1.1] - 2026-10-27

### Added

- An optional-features section in **Settings → About**. It reports the
  availability of DankSearch, thumbnail helpers, PDF preview, network shares,
  and removable drives, with manual installation commands where applicable.
- DankSearch now distinguishes **Not installed**, **Installed**, and
  **Active**; the duplicate, misleading index capability entry was removed.
- `dsearch-bin` is documented as the recommended AUR package for indexed,
  recursive search.
- Missing optional components now show a compact installation panel with a
  copyable command and a note to restart Omarchy Files before rechecking.

### Changed

- Document `qt6-webengine` as the optional dependency that enables in-app PDF
  preview.
- Preserve the version tag when deriving Arch development package versions.
- Strip local build paths from Arch packages and release binaries.

### Fixed

- Open the file-operation progress popover automatically for transfers that
  take longer than a brief moment, so copying and moving are visibly active.
- Complete copy operations correctly on NFS mounts that reject Linux
  `renameat2` semantics.
- Treat files on network mounts safely: delete permanently only after a clear
  confirmation when no trash is available, and return open tabs to Home when a
  remote location disconnects.

## [0.1.0] - 2026-10-23

### Added

- First Omarchy Files release: an Omarchy-native Wayland file manager based on
  Dux Files, with branded Arch packaging, a user service, and a live
  Omarchy-theme adapter.

[0.1.1]: https://github.com/spalencsar/omarchy-files/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/spalencsar/omarchy-files/releases/tag/v0.1.0
