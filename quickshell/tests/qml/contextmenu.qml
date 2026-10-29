pragma ComponentBehavior: Bound

import QtQuick
import Quickshell
import qs.Common
import qs.Modules
import qs.Services

ShellRoot {
    id: fixture

    readonly property var log: Log.scoped("fixture")
    property int failures: 0

    function check(label, actual, expected) {
        const ok = JSON.stringify(actual) === JSON.stringify(expected);
        if (!ok) {
            failures++;
            log.error(`FAIL ${label}: got ${JSON.stringify(actual)} want ${JSON.stringify(expected)}`);
            return;
        }
        log.info(`ok ${label}`);
    }

    Item {
        id: anchor

        width: 800
        height: 600
    }

    FileContextMenu {
        id: itemMenu

        anchorItem: anchor
        selectionCount: 1
    }

    Component.onCompleted: {
        const folderLink = itemMenu.entries.find(entry => entry.id === "folderActions");
        check("item menu links to current folder actions", {
            "label": folderLink?.label,
            "submenu": folderLink?.submenu,
            "visible": folderLink?.visible
        }, {
            "label": "Current folder",
            "submenu": true,
            "visible": true
        });

        itemMenu.inResults = true;
        const searchFolderLink = itemMenu.entries.find(entry => entry.id === "folderActions");
        check("cross-folder results hide the current folder link", searchFolderLink?.visible, false);

        itemMenu.forFolder = true;
        check("folder menu contains new folder", itemMenu.entries.some(entry => entry.id === "newFolder"), true);
        check("folder menu contains paste", itemMenu.entries.some(entry => entry.id === "paste"), true);

        if (failures === 0)
            log.info("FIXTURE PASS");
        Qt.callLater(Qt.quit);
    }
}
