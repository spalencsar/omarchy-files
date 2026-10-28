import QtQuick
import Quickshell.Io

FileView {
    id: root

    property Process _permissionFixer: Process {
        command: ["chmod", "600", "--", root.path]
    }

    onSaved: _permissionFixer.running = true
}
