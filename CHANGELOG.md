# Changelog

All notable changes to Omarchy Files are documented here.

## [Unreleased]

### Added

- Git-aware folders show the current branch and change count, plus compact
  status badges for modified, added, untracked, deleted, renamed, and
  conflicted files. Parent folders aggregate changes below them.
- Safe repository actions in the file and background context menus: stage or
  unstage selected paths, stage or unstage all changes, copy repository-relative
  paths, and open the repository root in a terminal. Working-tree content is
  never discarded, and selected paths are passed to Git through standard input
  instead of process arguments.

### Changed

- Restructure the sidebar around clear separator-based groups: fixed places,
  personal bookmarks, then drives and network locations. Device paths now live
  in tooltips, duplicate mounted bookmarks are hidden, and mount/disconnect
  actions appear at the trailing edge on hover or keyboard focus.

### Fixed

- Keep current-folder actions reachable when a dense list leaves no empty
  background to right-click. Every item menu now links to the folder menu for
  creating folders or documents, pasting, reloading, terminal actions, and
  other operations that target the folder being viewed.

## [0.1.3] - 2026-10-28

### Fixed

- Keep file names, source/destination paths, and raw backend errors out of
  desktop-notification process arguments. Transfer notifications now use
  generic completion, cancellation, and failure text; details remain inside
  Omarchy Files.

## [0.1.2] - 2026-10-28

### Added

- A keyboard-accessible hover detail card for the persistent transfer
  indicator. It shows the current item, source and destination, progress,
  transfer speed, and remaining time without opening the full job list.
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

[Unreleased]: https://github.com/spalencsar/omarchy-files/compare/v0.1.3...HEAD
[0.1.3]: https://github.com/spalencsar/omarchy-files/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/spalencsar/omarchy-files/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/spalencsar/omarchy-files/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/spalencsar/omarchy-files/releases/tag/v0.1.0
