import QtQuick
import qs.Common
import qs.DCommon.Widgets

Rectangle {
    id: root
    property string version: ""

    width: parent?.width ?? 0
    height: 96
    radius: Theme.cornerRadiusS
    color: Theme.surfaceContainer
    border.width: Theme.outlineWidth
    border.color: Theme.outlineVariant

    Image {
        id: appIcon
        anchors.left: parent.left
        anchors.leftMargin: Theme.spacingL
        anchors.verticalCenter: parent.verticalCenter
        width: 56
        height: 56
        source: Qt.resolvedUrl("../../assets/omarchy-files.svg")
        sourceSize.width: width
        sourceSize.height: height
        fillMode: Image.PreserveAspectFit
    }

    Column {
        anchors.left: appIcon.right
        anchors.leftMargin: Theme.spacingM
        anchors.right: parent.right
        anchors.rightMargin: Theme.spacingL
        anchors.verticalCenter: parent.verticalCenter
        spacing: 2

        StyledText { text: I18n.tr("Omarchy Files", "about page product name"); color: Theme.surfaceText; font.pixelSize: Theme.fontSizeXLarge; font.weight: Theme.fontWeightMedium }
        StyledText { text: I18n.tr("Standalone Wayland file manager for Omarchy", "about page product description"); color: Theme.surfaceVariantText; font.pixelSize: Theme.fontSizeSmall }
        StyledText { text: root.version; color: Theme.surfaceVariantText; font.pixelSize: Theme.fontSizeSmall }
        StyledText { text: I18n.tr("Maintained by Sebastian Palencsar · based on Dux Files", "about page maintainer and upstream credit"); color: Theme.surfaceVariantText; font.pixelSize: Theme.fontSizeSmall }
    }
}
