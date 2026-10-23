pragma ComponentBehavior: Bound

import QtQuick
import QtQuick.Layouts
import qs.Common
import qs.DCommon.Common
import qs.DCommon.FileBrowser
import qs.DCommon.Widgets

// App-owned sidebar row: intentionally plain, outlined and compact.  DCommon's
// stock FilePlaceRow uses a tinted circular icon well and a Material selection
// container, neither of which belongs in Omarchy Files.
StyledButton {
    id: row

    property string label: ""
    property string subtitle: ""
    property string tooltipText: ""
    property string iconName: ""
    property string glyph: ""
    property string actionIcon: ""
    property string actionTooltip: ""
    property real ringValue: -1
    property bool busy: false
    property bool selected: false
    property string dropPath: ""
    property bool springLoad: true

    signal actionTriggered
    signal dropped(var drop)
    signal springLoaded
    signal middleClicked

    readonly property bool showAction: actionIcon !== "" && (hovered || actionArea.containsMouse)
    readonly property color contentColor: !enabled ? Theme.onSurface_38 : selected ? Theme.onSelectedContainer : Theme.surfaceText

    focusPolicy: Qt.TabFocus
    width: parent?.width ?? 0
    implicitHeight: Math.max(Theme.buttonHeightXS + Theme.spacingXS, labels.implicitHeight + Theme.spacingS)
    radius: Theme.cornerRadiusS
    color: dropZone.containsDrag ? Theme.primaryContainer : selected ? Theme.surfaceContainerLow : "transparent"
    border.width: selected ? Theme.outlineWidth : 0
    border.color: selected ? Theme.outlineVariant : "transparent"
    Accessible.name: label
    Accessible.description: tooltipText

    FileDropArea {
        id: dropZone
        anchors.fill: parent
        dropPath: row.dropPath
        enabled: row.dropPath !== ""
        springLoad: row.springLoad && !row.selected
        onDropped: drop => row.dropped(drop)
        onSpringLoaded: row.springLoaded()
    }

    TapHandler {
        acceptedButtons: Qt.MiddleButton
        onTapped: row.middleClicked()
    }

    StateLayer { control: row; tooltipText: actionArea.containsMouse && row.actionTooltip !== "" ? "" : row.tooltipText; tooltipMultiline: true; stateColor: row.contentColor }
    FocusRing { visible: row.visualFocus; anchors.margins: Theme.focusRingWidth / 2; radius: Math.max(0, row.radius - Theme.focusRingWidth / 2) }

    RowLayout {
        anchors.fill: parent
        anchors.leftMargin: Theme.spacingS
        anchors.rightMargin: Theme.spacingS
        spacing: Theme.spacingS

        Item {
            Layout.preferredWidth: Theme.iconSizeMedium
            Layout.preferredHeight: Theme.iconSizeMedium

            DSpinner { anchors.centerIn: parent; visible: row.busy; size: Theme.iconSizeSmall; color: Theme.primary }
            FileIcon { anchors.centerIn: parent; visible: !row.busy && !row.showAction; iconName: row.iconName; glyph: row.glyph; size: Theme.iconSizeMedium; color: row.selected ? Theme.primary : Theme.surfaceVariantText }
            DIcon { anchors.centerIn: parent; visible: !row.busy && row.showAction; name: row.actionIcon; size: Theme.iconSizeMedium; color: Theme.primary }

            MouseArea {
                id: actionArea
                anchors.fill: parent
                enabled: row.actionIcon !== "" && !row.busy
                hoverEnabled: true
                cursorShape: Qt.PointingHandCursor
                onClicked: row.actionTriggered()
            }
        }

        ColumnLayout {
            id: labels
            Layout.fillWidth: true
            spacing: 0
            StyledText { Layout.fillWidth: true; text: row.label; color: row.contentColor; font.pixelSize: Theme.fontSizeMedium; wrapMode: Text.NoWrap; elide: Text.ElideMiddle }
            StyledText { Layout.fillWidth: true; visible: row.subtitle !== ""; text: row.subtitle; color: row.selected || !row.enabled ? row.contentColor : Theme.surfaceVariantText; font.pixelSize: Theme.fontSizeSmall; wrapMode: Text.NoWrap; elide: Text.ElideMiddle }
        }
    }
}
