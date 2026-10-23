import QtQuick
import qs.Common
import qs.DCommon.Widgets

Row {
    id: root
    property var model: []
    property int currentIndex: -1
    property int maximumWidth: -1
    property bool enabled: true
    // Keep a physical gutter in the delegate width. Some QML positioners can
    // otherwise collapse their inherited spacing property at runtime.
    property int optionGap: 6
    spacing: 0
    signal selectionChanged(int index, bool selected)

    Repeater {
        model: root.model
        delegate: Item {
            required property int index
            required property var modelData
            width: option.width + (index < root.model.length - 1 ? root.optionGap : 0)
            height: option.height

            Rectangle {
                id: option
                readonly property bool selected: index === root.currentIndex
                width: Math.max(56, optionText.implicitWidth + Theme.spacingM * 2)
                height: Theme.buttonHeightXS
                radius: Theme.cornerRadiusS
                color: selected ? Theme.surfaceContainerLow : Theme.surfaceContainerLowest
                border.width: Theme.outlineWidth
                border.color: selected ? Theme.primary : Theme.outlineVariant

                StyledText { id: optionText; anchors.centerIn: parent; text: typeof modelData === "object" ? modelData.label : modelData; color: option.selected ? Theme.primary : Theme.surfaceVariantText; font.pixelSize: Theme.fontSizeSmall }
                MouseArea { anchors.fill: parent; enabled: root.enabled; cursorShape: Qt.PointingHandCursor; onClicked: root.selectionChanged(index, true) }
            }
        }
    }
}
