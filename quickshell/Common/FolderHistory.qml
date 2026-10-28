pragma Singleton
pragma ComponentBehavior: Bound

import QtQuick
import Quickshell
import Quickshell.Io
import qs.Services
import "FolderJump.js" as FolderJump

Singleton {
    id: root

    readonly property string historyPath: Paths.strip(Paths.stateDir) + "/folders.json"
    readonly property string home: Paths.strip(Paths.home)

    property bool loaded: false
    property var entries: []
    property var verifiedAt: ({})
    readonly property var cache: ({
            "entries": null,
            "prepared": []
        })

    function prepared() {
        if (cache.entries !== entries) {
            cache.prepared = FolderJump.prepare(entries, home);
            cache.entries = entries;
        }
        return cache.prepared;
    }

    function load() {
        if (loaded)
            return true;
        if (!wanted)
            return false;
        historyFile.path = historyPath;
        const parsed = FolderJump.parse(historyFile.text() ?? "");
        entries = parsed ?? FolderJump.seed(SessionData.legacyRecentDirs, Date.now());
        SessionData.legacyRecentDirs = [];
        loaded = true;
        if (parsed === null && entries.length > 0)
            save();
        return true;
    }

    function unload() {
        if (!loaded)
            return;
        if (writeDebounce.running) {
            writeDebounce.stop();
            save();
        }
        loaded = false;
        entries = [];
        verifiedAt = {};
        cache.entries = null;
        cache.prepared = [];
        historyFile.path = "";
    }

    function save() {
        historyFile.setText(JSON.stringify({
            "version": 1,
            "folders": entries
        }));
    }

    function change(next) {
        entries = next;
        writeDebounce.restart();
    }

    function visit(path) {
        if (!path.startsWith("/") || !load())
            return;
        change(FolderJump.visit(entries, stored(path), Date.now()));
    }

    function forget(path) {
        if (!load())
            return;
        change(FolderJump.forget(entries, [path]));
    }

    function verify(paths) {
        const now = Date.now();
        const due = paths.filter(path => path.startsWith("/") && now - (verifiedAt[path] ?? 0) >= FilesMetrics.folderVerifyInterval);
        if (due.length === 0)
            return;
        due.forEach(path => verifiedAt[path] = now);
        DirectoryService.missingPaths(due, missing => {
            if (missing.length === 0 || !root.loaded)
                return;
            root.change(FolderJump.forget(root.entries, missing));
        });
    }

    function stored(path) {
        return FolderJump.storedPath(path, NetworkService.networkMounts);
    }

    function top(limit, exclude) {
        return FolderJump.top(prepared(), Date.now(), limit, stored(exclude ?? ""));
    }

    readonly property bool wanted: WindowService.count > 0 || ChooserService.count > 0

    onWantedChanged: {
        if (wanted) {
            load();
            return;
        }
        unload();
    }

    Timer {
        id: writeDebounce

        interval: FilesMetrics.folderHistoryWriteDelay
        onTriggered: root.save()
    }

    PrivateFileView {
        id: historyFile

        blockLoading: true
        atomicWrites: true
        printErrors: false
    }
}
