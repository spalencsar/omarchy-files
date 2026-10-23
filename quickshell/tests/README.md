# Shell tests

`make test-qml` runs two suites; `python3 quickshell/tests/run-qml.py logic` or `shell` runs one.

- `logic`: `*.test.mjs` under `node --test`. Pure functions are lifted out of the QML sources with `qml-script.mjs`, so they run without Qt. Put a test here when the code under test does not need a QML object tree.
- `shell`: every `qml/*.qml` is loaded as the `shell.qml` of a scratch root made of symlinks to the repo tree, in its own offscreen Quickshell. A fixture drives real singletons through a fake transport and logs `FIXTURE PASS` once every check passed.

Fixtures never see the host: `HOME` and every `XDG_*` dir point into a temp root, `OMARCHY_FILES_*`, `QS_*` and `QT_*` are dropped, the locale is `C.UTF-8` and the timezone UTC, no socket and no session bus is used. Each fixture runs in well under a second; a run that takes longer is a bug.

Rules: a test asserts behaviour a user would notice (batch application, ordering, error mapping), never structure. No sleeps, no timers waiting on the system, no network, no real files outside the temp root.

Needs Python 3, Node.js and Quickshell. Set `QS` to point at another quickshell binary.
