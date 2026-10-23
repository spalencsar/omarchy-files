import QtQuick
import qs.Common
import qs.DCommon.Widgets

DActionButton {
    focusPolicy: Qt.TabFocus
    buttonSize: FilesMetrics.topbarControlSize
    iconSize: Theme.iconSizeMedium
    iconColor: Theme.surfaceText
    circular: false
    radius: Theme.cornerRadiusS
    backgroundColor: Theme.surfaceContainerLowest
    border.width: Theme.outlineWidth
    border.color: Theme.outlineVariant
}
