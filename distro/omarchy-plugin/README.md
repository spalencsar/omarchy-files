# Omarchy Files for the Omarchy bar

This plugin adds an Omarchy Files launcher to the Omarchy bar. A click opens
the installed file manager. If it is not installed yet, the same click opens a
visible terminal installer for the pinned Omarchy Files v@APP_VERSION@ release.

![Omarchy Files browsing a sterile demo workspace](preview.png)

Omarchy Files is an independent, Omarchy-native evolution of
[Dux Files](https://github.com/duxlabs-dev/dux-files). The application remains
in its own [source repository](https://github.com/spalencsar/omarchy-files).
This Marketplace repository is generated from that canonical source and
contains only the shell launcher and optional user-local installer.

## Install the plugin

```bash
omarchy plugin add https://github.com/spalencsar/omarchy-files-omarchy --enable
```

The plugin recognizes Omarchy Files installed in `~/.local`, `/usr`, or
`/usr/local`. When the application is missing, click the widget and review the
installation summary in the terminal before confirming it.

## What the installer does

The installer:

- downloads the immutable Omarchy Files v@APP_VERSION@ release for x86_64 or aarch64;
- verifies it against the SHA-256 checksum pinned in the script;
- installs the binary, desktop entry, and icon below `~/.local`; and
- does not use `sudo`, enable a service, change MIME defaults, or enable the
  file chooser portal.

The user-local installation needs Omarchy, Quickshell, `curl`, `gzip`, and
standard core utilities. Those are present on a normal Omarchy installation.
Optional application integrations such as GVfs network locations, thumbnails,
and DankSearch remain separate and are explained in the
[application README](https://github.com/spalencsar/omarchy-files#installation-on-omarchy).

For full system package integration, including the polkit policy, D-Bus
activation, and optional user service, follow the Arch package instructions in
the application repository instead.

## Remove

Remove only the bar plugin:

```bash
omarchy plugin remove io.github.spalencsar.omarchy-files
```

That intentionally leaves the application installed. To remove the user-local
application installed by this plugin, run `uninstall-app.sh` from a checked-out
copy of this repository. It removes only the binary, desktop entry, and icon
installed by `install-omarchy.sh`; user settings are preserved.

If Omarchy Files was installed as an Arch package, remove it through pacman
instead. The local uninstall script does not touch system packages.

## Security and permissions

Omarchy plugins execute with the user's permissions. This plugin only checks
three fixed executable paths and launches either Omarchy Files or its visible
installer. Network access happens only after the user confirms installation.
Review `install-omarchy.sh` before accepting the download.

## Development

The source of this generated repository lives in
[`distro/omarchy-plugin`](https://github.com/spalencsar/omarchy-files/tree/master/distro/omarchy-plugin).
Changes should be made there, not in the generated Marketplace checkout.

```bash
omarchy plugin validate .
qmllint -I "$OMARCHY_PATH/shell" BarWidget.qml
```

The permanent plugin ID is `io.github.spalencsar.omarchy-files`.

## License

MIT. See [LICENSE](LICENSE). The Omarchy Files application is based on Dux
Files and retains its upstream MIT license and attribution in the application
repository and distributions.
