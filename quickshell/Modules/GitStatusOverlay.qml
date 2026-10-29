pragma ComponentBehavior: Bound

import QtQuick
import qs.Common
import qs.DCommon.FileBrowser

Item {
    id: overlay

    required property var pane
    property var state: ({})
    property int epoch: 0

    readonly property var entries: state?.repository === true ? (state.entries ?? []) : []
    readonly property var targetView: pane.body?.view?.activeList ?? pane.body?.view ?? null
    readonly property bool usable: !pane.resultsActive && !pane.virtualActive && targetView !== null

    function badgeRect(path, size) {
        const tick = epoch;
        if (!usable)
            return Qt.rect(0, 0, 0, 0);
        const index = pane.body.directory.indexOfPath(path);
        if (index < 0)
            return Qt.rect(0, 0, 0, 0);
        const item = targetView.itemAtIndex(index);
        if (!item || !item.visible)
            return Qt.rect(0, 0, 0, 0);
        const point = item.mapToItem(overlay, 0, 0);
        const iconSize = pane.body.iconSize;
        if (pane.viewMode === "grid") {
            return Qt.rect(point.x + (item.width - iconSize) / 2 + iconSize - size * 0.65,
                point.y + FileBrowserMetrics.gridTileVerticalPadding - size * 0.35, size, size);
        }
        return Qt.rect(point.x + FileBrowserMetrics.listRowPadding + iconSize - size * 0.65,
            point.y + (item.height - iconSize) / 2 - size * 0.35, size, size);
    }

    anchors.fill: parent
    visible: usable && entries.length > 0
    z: 20

    Repeater {
        model: overlay.entries

        GitStatusBadge {
            id: badge

            required property var modelData
            readonly property rect position: overlay.badgeRect(modelData.path, badgeSize)

            x: position.x
            y: position.y
            visible: position.width > 0
            code: modelData.code
            aggregate: modelData.aggregate === true
            badgeSize: Math.max(14, Math.min(18, overlay.pane.body.iconSize * 0.42))
        }
    }

    Connections {
        target: overlay.targetView
        ignoreUnknownSignals: true

        function onContentXChanged() { overlay.epoch++; }
        function onContentYChanged() { overlay.epoch++; }
        function onWidthChanged() { overlay.epoch++; }
        function onHeightChanged() { overlay.epoch++; }
        function onCountChanged() { overlay.epoch++; }
    }
}
