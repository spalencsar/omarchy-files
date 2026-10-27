# Omarchy Files in the Plugin Marketplace

Status: plan, 2026-10-26. No plugin or marketplace submission exists yet.

## Intended experience

The marketplace entry gives Omarchy users a small Files launcher in the
Omarchy shell. A click opens the installed Omarchy Files application. If the
application is missing, the launcher offers a visible installation path in a
terminal. Installing or enabling the plugin must not silently change the
default file manager, enable a service, or select a file chooser portal.

This follows the BearWave Omarchy pattern: a shell plugin provides the desktop
entry point while the application keeps its own process and lifecycle. Omarchy
plugin installation clones and validates a repository; it does not run package
installation hooks.

## 1. Prepare a reproducible application build

- Cut a new tested Omarchy Files release. The existing v0.1.0 tag predates
  subsequent file-operation and UI fixes.
- Make the Arch recipe used by the plugin installer build an exact release
  commit, with the source revision and external archive checksums recorded.
  The current development PKGBUILD tracks Git HEAD; that is unsuitable for a
  repeatable marketplace installation.
- In a visible terminal, let the user confirm the build dependencies, package
  build, and package installation. The plugin must not request privileges or
  run a package installation simply because it was added or enabled.
- Give application updates an explicit path. Updating the small shell plugin
  must not silently replace the installed file manager; show the installed
  and offered application versions before a confirmed upgrade.
- Verify installation, launch, service behavior, and removal on a clean
  Omarchy session. Preserve user data on uninstall unless the user explicitly
  chooses to delete it.

Exit condition: a user who adds the plugin can install the tested application
through the visible first-run flow, without following manual build steps.

## 2. Build a small, separate Omarchy plugin

- Use a dedicated public plugin repository for the shell integration and
  first-run installer, following the BearWave Omarchy layout. Keep the app's
  source and release history in the existing Omarchy Files repository.
- Add a root `manifest.json` with schema version 1, a unique lowercase
  namespaced ID, a version, author, description, `bar-widget` kind, and a
  matching QML entry point. Check the proposed ID against the marketplace
  before settling on it.
- Make the bar widget show a restrained Files icon and a clear tooltip. On
  click, launch the installed desktop application. When the package is absent,
  open a visible terminal with the pinned-release installer and ask the user
  to confirm each system-changing step.
- Keep checks lightweight. The widget should detect installation and recover
  after an install or update without forcing an Omarchy shell restart.
- Document that removing the plugin removes the launcher, while removing the
  application is a separate, explicit package action. Describe the optional
  user-service and default-manager choices separately.
- Include a README, MIT license and upstream attribution, dependency list,
  install and removal instructions, and a preview without private folder
  names or paths.

Exit condition: adding the plugin leaves a usable widget even before the
application is installed, and removing it has no hidden effect on files,
defaults, or the installed package.

## 3. Validate the complete path

- Run `omarchy plugin validate` and `qmllint` against the plugin entry point
  on a compatible Omarchy shell.
- In a clean session, test: plugin add and enable; first click with no package;
  confirmed install; launch; theme switch while running; package upgrade;
  disabled plugin; plugin removal; and explicit application removal.
- Check coexistence with another file manager. Test `FileManager1` ownership
  and directory MIME association without changing user defaults automatically.
- Verify that failed package installation leaves the plugin usable and offers
  an understandable retry. Test the user service with and without prior
  enablement.

Exit condition: the plugin and application can each be installed, updated,
disabled, and removed without leaving a broken launcher or changing desktop
defaults unexpectedly.

## 4. Submit

- Check the public plugin repository against the publishing checklist: root
  `manifest.json`, existing QML entry point, README, license, documented
  dependencies, safe installation and removal, and an optional preview.
- Choose the marketplace category and one to three allowed tags; likely
  `System` with `quickshell` and `system`, subject to the current form.
- Submit the plugin repository through the marketplace issue form only after
  the release and clean-session checks pass. Automated validation checks the
  submitted commit; a maintainer must approve the listing. Review any
  findings before publication. Listing is not a security certification.
- Link the listing from the Omarchy Files README after it is approved.

## Later: simplify installation through the Omarchy package repository

The marketplace route above does not depend on a package-repository
submission. Once the integration works, propose a release-based package to
`omacom/omarchy-pkgs`. That repository requires immutable sources,
architecture-specific clean builds, and `.omarchy/package.json` metadata;
the current `-git` PKGBUILD cannot be copied over unchanged. Once a package
is published for users' Omarchy channels, the plugin can replace its local
build flow with the supported `omarchy pkg add` command.

## References

- BearWave's Omarchy plugin: https://github.com/spalencsar/bearwave_omarchy
- Marketplace publishing guide: https://plugins.omarchy.org/publish.html
- Omarchy shell plugin contract:
  https://github.com/omacom/omarchy/blob/quattro/docs/omarchy-shell.md
- Marketplace submission rules:
  https://github.com/omacom/omarchy-plugin-marketplace/blob/main/SUBMISSION.md
- Omarchy package contribution rules:
  https://github.com/omacom/omarchy-pkgs
