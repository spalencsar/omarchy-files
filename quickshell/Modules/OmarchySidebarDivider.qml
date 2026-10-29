import QtQuick
import qs.Common

Item {
    width: parent?.width ?? 0
    implicitHeight: Theme.spacingL

    Rectangle {
        anchors.left: parent.left
        anchors.right: parent.right
        anchors.leftMargin: Theme.spacingS
        anchors.rightMargin: Theme.spacingS
        anchors.verticalCenter: parent.verticalCenter
        height: Theme.outlineWidth
        color: Theme.outlineVariant
    }
}
