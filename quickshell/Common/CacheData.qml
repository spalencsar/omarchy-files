pragma Singleton
pragma ComponentBehavior: Bound

import QtQuick
import Quickshell
import Quickshell.Io

Singleton {
    id: root

    property alias fileBrowserSettings: adapter.fileBrowserSettings
    property alias sidebarWidth: adapter.sidebarWidth
    property alias previewWidth: adapter.previewWidth

    function saveCache() {
        if (!cacheFile.loaded)
            return;
        writeDebounce.restart();
    }

    Timer {
        id: writeDebounce

        interval: 400
        onTriggered: {
            if (!Paths.ready) {
                restart();
                return;
            }
            cacheFile.writeAdapter();
        }
    }

    PrivateFileView {
        id: cacheFile

        property bool loaded: false

        path: Paths.strip(Paths.cache) + "/cache.json"
        blockLoading: true
        atomicWrites: true
        printErrors: false
        onLoaded: loaded = true
        onLoadFailed: loaded = true
        onAdapterUpdated: root.saveCache()

        JsonAdapter {
            id: adapter

            property var fileBrowserSettings: ({})
            property int sidebarWidth: FilesMetrics.sidebarWidth
            property int previewWidth: 0
        }
    }
}
