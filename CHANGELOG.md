# Changelog

All notable changes to Omarchy Files are documented here.

## Unreleased

### Added

- An optional-features section in **Settings → About**. It reports the
  availability of DankSearch, thumbnail helpers, PDF preview, network shares,
  and removable drives, with manual installation commands where applicable.
- DankSearch now distinguishes **Not installed**, **Installed**, and
  **Active**; the duplicate, misleading index capability entry was removed.
- `dsearch-bin` is documented as the recommended AUR package for indexed,
  recursive search.

### Changed

- Document `qt6-webengine` as the optional dependency that enables in-app PDF
  preview.
- Preserve the version tag when deriving Arch development package versions.

### Fixed

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

[0.1.0]: https://github.com/spalencsar/omarchy-files/releases/tag/v0.1.0
