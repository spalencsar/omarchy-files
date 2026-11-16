pragma ComponentBehavior: Bound

import QtQuick
import Quickshell
import qs.Common
import qs.Modules
import qs.Services
import qs.DCommon.FileBrowser

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

    function file(size) {
        return {
            "isDir": false,
            "size": size
        };
    }

    function folder() {
        return {
            "isDir": true,
            "size": 4096
        };
    }

    PaneStatus {
        id: status

        total: 7
        hidden: 0
        compact: false
        showClipboard: false
    }

    Component.onCompleted: {
        check("folder total alone", status.statusText, "7 items");

        status.hidden = 2;
        check("hidden count joins the total", status.statusText, "7 items · 2 hidden");

        status.hidden = 0;
        status.selectedEntries = [file(1200)];
        check("one selected file adds count and size", status.statusText, "7 items · 1 selected · " + FileFormat.size(1200));

        status.selectedEntries = [file(4000000), file(8400000), folder()];
        check("mixed selection sums files and skips folder bytes", status.statusText, "7 items · 3 selected · " + FileFormat.size(12400000));

        status.selectedEntries = [folder(), folder()];
        check("folders alone omit a size", status.statusText, "7 items · 2 selected");

        status.selectedEntries = [file(-1), file(12)];
        check("unknown file size omits the size", status.statusText, "7 items · 2 selected");

        status.compact = true;
        status.selectedEntries = [file(1500), file(500)];
        check("compact windows keep the selection and drop the folder total", status.statusText, "2 selected · " + FileFormat.size(2000));

        status.selectedEntries = [];
        check("compact windows hide the folder total", status.statusText, "");

        if (failures === 0)
            log.info("FIXTURE PASS");
        Qt.callLater(Qt.quit);
    }
}
