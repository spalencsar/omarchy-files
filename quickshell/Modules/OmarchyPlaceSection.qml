pragma ComponentBehavior: Bound

import QtQuick
import qs.Common
import qs.DCommon.Widgets
import qs.DCommon.FileBrowser

Column {
    id: section

    property string title: ""
    property string emptyText: ""
    property bool showEmpty: false
    property bool divider: false

    default property alias rows: body.data

    spacing: FileBrowserMetrics.sidebarRowGap

    OmarchySidebarDivider {
        visible: section.divider
    }

    StyledText {
        visible: section.title !== ""
        text: section.title
        color: Theme.surfaceVariantText
        font.pixelSize: Theme.fontSizeSmall
        font.weight: Theme.fontWeightMedium
        leftPadding: Theme.spacingS
        bottomPadding: Theme.spacingXXS
    }

    Column {
        id: body

        width: section.width
        spacing: FileBrowserMetrics.sidebarRowGap
    }

    StyledText {
        width: section.width
        visible: section.showEmpty && section.emptyText !== ""
        text: section.emptyText
        color: Theme.surfaceVariantText
        font.pixelSize: Theme.fontSizeSmall
        leftPadding: Theme.spacingS
        rightPadding: Theme.spacingS
        wrapMode: Text.Wrap
    }
}
