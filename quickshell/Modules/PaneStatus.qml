pragma ComponentBehavior: Bound

import QtQuick
import qs.Common
import qs.Services
import qs.DCommon.Widgets
import qs.DCommon.FileBrowser

Row {
    id: status

    property int total: 0
    property int hidden: 0
    property var selectedEntries: []
    property bool compact: false
    property bool showClipboard: true
    property bool gitRepository: false
    property string gitBranch: ""
    property int gitChanged: 0

    readonly property int clipboardCount: ClipboardService.paths.length
    readonly property int selectedCount: selectedEntries?.length ?? 0
    readonly property real selectedBytes: {
        const entries = selectedEntries ?? [];
        let totalBytes = 0;
        let files = 0;
        for (const entry of entries) {
            if (entry?.isDir)
                continue;
            if ((entry?.size ?? -1) < 0)
                return -1;
            totalBytes += entry.size;
            files++;
        }
        return files > 0 ? totalBytes : -1;
    }
    readonly property string countText: {
        const items = FileFormat.count(total);
        if (hidden <= 0)
            return items;
        return I18n.tr("%1 · %2", "folder status: item count then hidden count").arg(items).arg(I18n.tr("%1 hidden", "count of hidden items left out of a folder listing").arg(hidden));
    }
    readonly property string selectionText: {
        if (selectedCount <= 0)
            return "";
        const items = selectedCount === 1 ? I18n.tr("1 selected", "status line for a single selected item") : I18n.tr("%1 selected", "status line count of selected items").arg(selectedCount);
        if (selectedBytes < 0)
            return items;
        return I18n.tr("%1 · %2", "folder status: selection count then size").arg(items).arg(FileFormat.size(selectedBytes));
    }
    readonly property string statusText: {
        const showCount = !compact && (total > 0 || hidden > 0);
        if (showCount && selectionText !== "")
            return I18n.tr("%1 · %2", "folder status: item count then selection").arg(countText).arg(selectionText);
        if (showCount)
            return countText;
        return selectionText;
    }
    readonly property string clipboardText: {
        switch (true) {
        case ClipboardService.cut && clipboardCount === 1:
            return I18n.tr("1 item cut", "clipboard indicator for a single cut item");
        case ClipboardService.cut:
            return I18n.tr("%1 items cut", "clipboard indicator for cut items").arg(clipboardCount);
        case clipboardCount === 1:
            return I18n.tr("1 item copied", "clipboard indicator for a single copied item");
        default:
            return I18n.tr("%1 items copied", "clipboard indicator for copied items").arg(clipboardCount);
        }
    }

    spacing: Theme.spacingS

    StyledText {
        anchors.verticalCenter: parent.verticalCenter
        visible: status.statusText !== ""
        text: status.statusText
        color: Theme.surfaceVariantText
        font.pixelSize: Theme.fontSizeSmall
        Accessible.role: Accessible.StaticText
        Accessible.name: status.statusText
    }

    Rectangle {
        anchors.verticalCenter: parent.verticalCenter
        visible: status.gitRepository
        width: gitChip.implicitWidth + FileBrowserMetrics.pathPillPadding * 2
        height: FileBrowserMetrics.pathPillHeight
        radius: FileBrowserMetrics.pathPillRadius
        color: Theme.chipSurface
        border.width: Theme.outlineWidth
        border.color: status.gitChanged > 0 ? Theme.warning : Theme.outlineVariant
        Accessible.role: Accessible.StaticText
        Accessible.name: I18n.tr("Git branch %1, %2 changed", "repository status: branch then number of changed paths").arg(status.gitBranch).arg(status.gitChanged)

        Row {
            id: gitChip

            anchors.centerIn: parent
            spacing: FileBrowserMetrics.pathPillSpacing

            DIcon {
                anchors.verticalCenter: parent.verticalCenter
                name: "account_tree"
                size: FileBrowserMetrics.pathIconSize
                color: status.gitChanged > 0 ? Theme.warning : Theme.primary
            }

            StyledText {
                anchors.verticalCenter: parent.verticalCenter
                text: status.compact ? String(status.gitChanged) : I18n.tr("%1 · %2", "compact Git status: branch then change count").arg(status.gitBranch).arg(status.gitChanged)
                color: Theme.surfaceText
                font.pixelSize: Theme.fontSizeSmall
                font.weight: Theme.fontWeightMedium
            }
        }
    }

    Rectangle {
        anchors.verticalCenter: parent.verticalCenter
        visible: status.showClipboard && status.clipboardCount > 0
        width: chip.implicitWidth + FileBrowserMetrics.pathPillPadding + Theme.spacingXXS
        height: FileBrowserMetrics.pathPillHeight
        radius: FileBrowserMetrics.pathPillRadius
        color: Theme.chipSurface
        border.width: Theme.outlineWidth
        border.color: Theme.outlineVariant
        Accessible.role: Accessible.StaticText
        Accessible.name: status.clipboardText

        Row {
            id: chip

            anchors.left: parent.left
            anchors.leftMargin: FileBrowserMetrics.pathPillPadding
            anchors.verticalCenter: parent.verticalCenter
            spacing: FileBrowserMetrics.pathPillSpacing

            DIcon {
                anchors.verticalCenter: parent.verticalCenter
                name: ClipboardService.cut ? "content_cut" : "content_copy"
                size: FileBrowserMetrics.pathIconSize
                color: Theme.primary
            }

            StyledText {
                anchors.verticalCenter: parent.verticalCenter
                text: status.compact ? String(status.clipboardCount) : status.clipboardText
                color: Theme.surfaceText
                font.pixelSize: Theme.fontSizeMedium
                font.weight: Theme.fontWeightMedium
            }

            DActionButton {
                anchors.verticalCenter: parent.verticalCenter
                buttonSize: FileBrowserMetrics.pathPillHeight - Theme.spacingXXS * 2
                iconName: "close"
                iconSize: Theme.iconSizeSmall
                iconColor: Theme.surfaceVariantText
                tooltipText: I18n.tr("Clear the clipboard", "tooltip on the button dropping the copied or cut items")
                onClicked: ClipboardService.clear()
            }
        }
    }
}
