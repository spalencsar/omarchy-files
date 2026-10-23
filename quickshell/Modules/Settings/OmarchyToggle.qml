import QtQuick
import qs.Common

// Compact state control for settings. It deliberately avoids the large,
// filled Material switch track and circular thumb.
Rectangle {
    id: root

    property bool checked: false
    property bool toggling: false
    property string text: ""
    property string description: ""
    property bool hideText: true

    signal toggled(bool checked)

    width: 42
    height: 22
    enabled: !toggling
    focusPolicy: Qt.ClickFocus
    radius: Theme.cornerRadiusS
    color: checked ? Theme.surfaceContainerLow : Theme.surfaceContainerLowest
    border.width: Theme.outlineWidth
    border.color: checked ? Theme.primary : Theme.outlineVariant
    Accessible.role: Accessible.CheckBox
    Accessible.name: text
    Accessible.description: description
    Accessible.checkable: true
    Accessible.checked: checked
    Accessible.onToggleAction: root.toggled(!root.checked)

    MouseArea {
        anchors.fill: parent
        enabled: root.enabled
        cursorShape: Qt.PointingHandCursor
        onClicked: root.toggled(!root.checked)
    }

    Rectangle {
        width: 12
        height: 12
        anchors.verticalCenter: parent.verticalCenter
        x: root.checked ? parent.width - width - Theme.spacingXS : Theme.spacingXS
        radius: Theme.cornerRadiusXS
        color: root.checked ? Theme.primary : Theme.surfaceVariantText

        Behavior on x {
            enabled: Theme.animationsEnabled
            NumberAnimation { duration: Theme.shorterDuration }
        }
    }
}
