pragma ComponentBehavior: Bound

import QtQuick
import Quickshell
import qs.Common
import qs.Services

ShellRoot {
    id: fixture

    readonly property var log: Log.scoped("fixture")
    property int failures: 0
    property int attempts: 0

    function check(label, actual, expected) {
        const ok = JSON.stringify(actual) === JSON.stringify(expected);
        if (!ok) {
            failures++;
            log.error(`FAIL ${label}: got ${JSON.stringify(actual)} want ${JSON.stringify(expected)}`);
        }
    }

    function finish() {
        if (failures === 0)
            log.info("FIXTURE PASS");
        Qt.callLater(Qt.quit);
    }

    Timer {
        interval: 100
        repeat: true
        running: true
        onTriggered: {
            attempts++;
            if (Theme.omarchyColorsAvailable) {
                check("the active Omarchy color file takes precedence", Theme.primary.toString(), "#336699");
                check("the active Omarchy background becomes the application surface", Theme.surface.toString(), "#101112");
                check("the active Omarchy selection maps to selected rows", Theme.selectedContainer.toString(), "#202122");
                check("the active Omarchy mode controls automatic light mode", Theme.isLightMode, true);
                finish();
                stop();
                return;
            }
            if (attempts >= 30) {
                check("the active Omarchy colors load", Theme.omarchyColorsAvailable, true);
                finish();
                stop();
            }
        }
    }
}
