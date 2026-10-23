# Contributing

Contributions are welcome and encouraged.

To contribute fork this repository, make your changes, and open a pull request.

## Setup

Clone with submodules. The shared widget library ([dux-qml-common](https://github.com/duxlabs-dev/dux-qml-common)) is vendored at `dux-qml-common/` and symlinked into `quickshell/DCommon`:

```bash
git clone --recurse-submodules https://github.com/spalencsar/omarchy-files.git
# or, in an existing clone:
git submodule update --init
```

Install [prek](https://prek.j178.dev/) then activate pre-commit hooks:

```bash
prek install
```

The hooks run `gofmt -s`, `go vet`, `go mod tidy`, and the Go test suite for changes under `core/`, plus generic whitespace/YAML checks.

## Running

```bash
make run
# or with hot reload of the QML tree
OMARCHY_FILES_ENABLE_HOTRELOAD=1 core/bin/omarchy-files run -c quickshell
```

`omarchy-files ipc list` prints every IPC method the daemon exposes.

## Shared widgets (dux-qml-common)

Everything under `quickshell/DCommon/` is shared across the DMS suite and lives in the `dux-qml-common` submodule. It is a normal git worktree:

1. Edit files under `dux-qml-common/` (or through the `quickshell/DCommon` symlink, same files) and test in the running app. For isolated widget work, the library is its own runnable config with a gallery: `qs -p dux-qml-common`.
2. Commit and PR those changes in the `dux-qml-common` repo.
3. Once merged, bump the pointer here with `make update-common` and commit the submodule change.

If a library change reads a new app-singleton property, add it to `quickshell/Common/Theme.qml` or `SettingsData.qml` in the same PR that bumps the pointer.

## Translations

See `quickshell/translations/README.md`.

## Generative AI

Using an LLM to help write code is fine. Submitting its output unread is not.

- You are responsible for every line you submit. You have read it, tested it, and can explain it in review.
- Say in the PR when a meaningful part of the code was AI generated.
- Write PR descriptions, issues, and comments yourself. No AI-written or AI-rewritten text. A few plain sentences in your own words beat a page of generated summary. PRs and issues with AI-written text get closed without review.
- Do not file issues or leave comments you have not verified yourself. Reports that do not reproduce get closed.
- PRs that read like unreviewed output, with narrating comments, invented APIs, or style that ignores the file they are in, get closed without review.
