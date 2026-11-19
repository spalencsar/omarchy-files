# Omarchy Files in the Plugin Marketplace

Status: delivered. Current application release is **v0.1.5** (2026-11-16).
The existing Marketplace listing still points to the generated wrapper while
maintainers review a repository URL migration. The main repository now has a
root plugin manifest and generated plugin files, refreshed with pinned release
hashes after each app release.

The Marketplace entry is live as a generated wrapper. Adding the plugin puts
an Omarchy Files launcher on the Omarchy bar. A click opens the installed
application. If it is missing, the same click opens a visible terminal
installer for the pinned release. Installing or enabling the plugin does not
change the default file manager, enable a service, or select a file chooser
portal.

This follows the BearWave Omarchy pattern: a shell plugin provides the desktop
entry point while the application keeps its own process and lifecycle. Omarchy
plugin installation clones and validates a repository; it does not run package
installation hooks.

## What shipped

- Canonical plugin sources live in `distro/omarchy-plugin/` in this repository.
- The main repository's root `manifest.json` points at the generated
  `marketplace/` plugin. `scripts/render-marketplace-plugin.sh` renders it from
  the canonical source and a release's SHA-256 assets; the release workflow
  commits that output to `master` after publishing the app release.
- The small public wrapper is
  [`spalencsar/omarchy-files-omarchy`](https://github.com/spalencsar/omarchy-files-omarchy).
  It exists so `omarchy plugin add` does not clone the full application and
  its submodule.
- `.github/workflows/sync.yml` in that repository reads the latest public app
  release, pins amd64 and arm64 SHA-256 sums, copies the canonical assets,
  validates the output, and commits only when something changed. It runs daily
  at 05:17 UTC and through `workflow_dispatch`.
- Wrapper version **1.1.0** (generated commit `daae196` first pinned
  application **v0.1.4**). After a tagged GitHub release the wrapper sync
  pins the new version and both SHA-256 sums. The bar widget compares a
  plugin-managed `~/.local` binary with that pinned release, shows an update
  badge when a newer one is available, and reopens the same
  confirmation-based installer. Installations under `/usr` and `/usr/local`
  stay under their package manager.
- Installer and uninstaller scripts stay user-visible. They ask before
  downloading, verify pinned hashes, install below `~/.local`, and never use
  `sudo` or change MIME, portal, or service state.

Install the plugin with:

```bash
omarchy plugin add https://github.com/spalencsar/omarchy-files-omarchy --enable
```

The application README documents this path as the usual Omarchy install route.
The Arch package remains the path for polkit, D-Bus activation, and the
optional user service.

## Security review

Marketplace issue `omacom/omarchy-plugin-marketplace#13098` holds the review.

- Plugin commit `9bb7019` still delivered v0.1.1; v0.1.2 plus plugin commit
  `be95382` closed that delivery gap.
- Notification process arguments leaked private transfer names and raw errors;
  v0.1.3 plus plugin commit `ecb0d8c` closed that issue (comment `6474468657`).
- The reviewer accepted those findings as resolved. The plugin is listed at
  https://plugins.omarchy.org/plugin.html?id=io.github.spalencsar.omarchy-files.
  Manual maintainer review of installer scripts remains expected on later
  wrapper changes.

Keep the security baseline: visible confirmation, immutable release hashes,
install below `~/.local`, no sudo, no MIME/portal/service changes.

## Remaining follow-up

- Keep the published listing current through the generated wrapper until the
  Marketplace maintainers migrate its registered repository URL. Do not delete,
  archive, or rename the wrapper before that migration is confirmed.
- Later: propose a release-based package to `omacom/omarchy-pkgs`. That
  repository needs immutable sources, architecture-specific clean builds, and
  `.omarchy/package.json` metadata; the current `-git` PKGBUILD cannot be
  copied over unchanged. Once a package is in users' Omarchy channels, the
  plugin can replace its local download flow with `omarchy pkg add`.

## References

- BearWave's Omarchy plugin: https://github.com/spalencsar/bearwave_omarchy
- Marketplace publishing guide: https://plugins.omarchy.org/publish.html
- Omarchy shell plugin contract:
  https://github.com/omacom/omarchy/blob/quattro/docs/omarchy-shell.md
- Marketplace submission rules:
  https://github.com/omacom/omarchy-plugin-marketplace/blob/main/SUBMISSION.md
- Omarchy package contribution rules:
  https://github.com/omacom/omarchy-pkgs
