# Arch package

`PKGBUILD` produces `omarchy-files-git`, a development package built from the
current `main` branch of this repository. It provides `omarchy-files` but does
not install it as the default file manager, enable its user service, or select
its portal backend.

## Build locally

```bash
cd distro/arch
makepkg -s
```

The build fetches this repository and the pinned `dux-qml-common` dependency.
It compiles the embedded Quickshell UI, runs the Go test suite, packages shell
completions, and stages desktop integration under standard Arch locations.

## Install deliberately

After inspecting the produced package, install it with:

```bash
sudo pacman -U omarchy-files-git-*.pkg.tar.zst
```

Installation alone does not change MIME defaults, portal preferences, or user
service state. Those integrations remain separate, reversible user actions.

## Package contents

- `/usr/bin/omarchy-files` and Bash, Zsh, Fish, and PowerShell completions
- desktop entry, AppStream metadata, and icon
- polkit policy plus FileManager1 and portal D-Bus activation files
- portal backend descriptor
- inactive systemd user unit at `/usr/lib/systemd/user/omarchy-files.service`

Optional thumbnail, device, network-share, and PDF-preview features are listed
as `optdepends` by the package.
