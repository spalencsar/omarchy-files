import QtQuick
import qs.Common
import qs.DCommon.Widgets

Row {
    id: root
    property var model: []
    property int currentIndex: -1
    property int maximumWidth: -1
    property bool enabled: true
    property real spacing: Theme.spacingXS
    signal selectionChanged(int index, bool selected)

    Repeater {
        model: root.model
        delegate: Rectangle {
            required property int index
            required property var modelData
            readonly property bool selected: index === root.currentIndex
            width: Math.max(56, optionText.implicitWidth + Theme.spacingM * 2)
            height: Theme.buttonHeightXS
            radius: Theme.cornerRadiusS
            color: selected ? Theme.surfaceContainerLow : Theme.surfaceContainerLowest
            border.width: Theme.outlineWidth
            border.color: selected ? Theme.primary : Theme.outlineVariant

            StyledText { id: optionText; anchors.centerIn: parent; text: typeof modelData === "object" ? modelData.label : modelData; color: selected ? Theme.primary : Theme.surfaceVariantText; font.pixelSize: Theme.fontSizeSmall }
            MouseArea { anchors.fill: parent; enabled: root.enabled; cursorShape: Qt.PointingHandCursor; onClicked: root.selectionChanged(index, true) }
        }
    }
}
