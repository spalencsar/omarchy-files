import QtQuick
import qs.Common
import qs.Services
import qs.DCommon.Widgets

DActionButton {
    id: button

    readonly property bool operationActive: FileOpsService.activeCount > 0
    property bool autoShownForRun: false

    visible: FileOpsService.jobs.length > 0
    iconName: operationActive ? "sync" : "check_circle"
    buttonSize: Theme.buttonHeightXXS
    iconSize: Theme.iconSizeSmall
    iconColor: Theme.surfaceText
    tooltipText: I18n.tr("File operations", "jobs popover header")
    onClicked: popover.visible ? popover.close() : popover.open()

    // Make longer operations visible without requiring the user to notice and
    // click the small toolbar icon. The delay avoids flashing the popover for
    // quick actions such as renaming a file.
    onOperationActiveChanged: {
        if (!operationActive)
            autoShownForRun = false;
    }

    Timer {
        id: autoOpen

        interval: 400
        repeat: false
        running: button.visible && button.operationActive && !button.autoShownForRun
        onTriggered: {
            button.autoShownForRun = true;
            if (button.operationActive && button.visible && !popover.visible)
                popover.open();
        }
    }

    JobsPopover {
        id: popover

        y: button.height + Theme.spacingXS
        x: button.width - width
    }
}
