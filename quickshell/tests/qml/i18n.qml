pragma ComponentBehavior: Bound

import QtQuick
import Quickshell
import qs.Common
import qs.Services
import qs.DCommon.Common as DC
import qs.DCommon.FileBrowser

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
            return;
        }
        log.info(`ok ${label}`);
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
            if (!("de" in I18n.presentLocales) && attempts < 30)
                return;
            if (SettingsData.language !== "de") {
                DC.I18n.backend = I18n;
                SettingsData.language = "de";
                return;
            }
            if (!I18n.translationsLoaded && attempts < 50)
                return;

            check("German catalog is listed", "de" in I18n.presentLocales, true);
            check("German catalog loaded", I18n.translationsLoaded, true);
            check("new folder", I18n.tr("New folder", "file browser toolbar button creating a folder"), "Neuer Ordner");
            check("one selected", I18n.tr("1 selected", "status line for a single selected item"), "1 ausgewählt");
            check("many selected", I18n.tr("%1 selected", "status line count of selected items").arg(3), "3 ausgewählt");
            check("trash", I18n.tr("Trash", "trash view header and sidebar entry"), "Papierkorb");
            check("item count uses DCommon strings from the app catalog", FileFormat.count(7), "7 Elemente");
            check("unknown terms stay in English", I18n.tr("Not a real string", "test"), "Not a real string");

            SettingsData.language = "en";
            check("English is still available", I18n.tr("New folder", "file browser toolbar button creating a folder"), "New folder");

            if (attempts >= 50 && !I18n.translationsLoaded)
                check("German catalog loaded", I18n.translationsLoaded, true);

            finish();
            stop();
        }
    }
}
