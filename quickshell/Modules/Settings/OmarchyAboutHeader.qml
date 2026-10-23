import QtQuick
import qs.Common
import qs.DCommon.Widgets

Rectangle {
    id: root
    property string version: ""
    signal githubRequested

    height: 116
    radius: Theme.cornerRadiusS
    color: Theme.surfaceContainer
    border.width: Theme.outlineWidth
    border.color: Theme.outlineVariant

    Column {
        anchors.centerIn: parent
        spacing: Theme.spacingS

        StyledText { anchors.horizontalCenter: parent.horizontalCenter; text: I18n.tr("Omarchy Files", "about page product name"); color: Theme.surfaceText; font.pixelSize: Theme.fontSizeXXLarge; font.weight: Theme.fontWeightMedium }
        StyledText { anchors.horizontalCenter: parent.horizontalCenter; text: root.version; color: Theme.surfaceVariantText; font.pixelSize: Theme.fontSizeSmall }

        Rectangle {
            width: actionLabel.implicitWidth + Theme.spacingL * 2
            height: Theme.buttonHeightXS
            anchors.horizontalCenter: parent.horizontalCenter
            radius: Theme.cornerRadiusS
            color: "transparent"
            border.width: Theme.outlineWidth
            border.color: Theme.outlineVariant
            StyledText { id: actionLabel; anchors.centerIn: parent; text: I18n.tr("View project on GitHub", "about page project link"); color: Theme.primary; font.pixelSize: Theme.fontSizeSmall }
            MouseArea { anchors.fill: parent; cursorShape: Qt.PointingHandCursor; onClicked: root.githubRequested() }
        }
    }
}
