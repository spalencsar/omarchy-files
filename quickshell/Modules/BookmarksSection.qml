pragma ComponentBehavior: Bound

import QtQuick
import qs.Common
import qs.Services
import qs.DCommon.Widgets
import qs.DCommon.FileBrowser

OmarchyPlaceSection {
    id: section

    property string currentPath: ""

    signal navigateRequested(string path)
    signal tabRequested(string path)
    signal dropRequested(string path, var drop)
    signal menuRequested(string path, real pointX, real pointY)
    signal trashMenuRequested(real pointX, real pointY)
    signal recentMenuRequested(real pointX, real pointY)

    readonly property string homePath: DirectoryService.userDirs.find(dir => dir.key === "home")?.path ?? ""
    readonly property var essentialPaths: [homePath, NetworkService.path, RecentService.path, TrashService.path].filter(path => path !== "")
    readonly property var essentials: essentialEntries()
    readonly property var bookmarks: savedEntries()

    function normalize(path) {
        return path?.length > 1 && path.endsWith("/") ? path.slice(0, -1) : path ?? "";
    }

    function mountedPaths() {
        const paths = [];
        for (const device of DevicesService.devices) {
            for (const path of device.mountPoints ?? [device.mountPoint]) {
                if (path)
                    paths.push(normalize(path));
            }
        }
        for (const mount of NetworkService.mounts) {
            if (mount.path)
                paths.push(normalize(mount.path));
        }
        return paths;
    }

    function essentialEntries() {
        const entries = [];
        for (const path of essentialPaths) {
            if (path === RecentService.path && !RecentService.shown)
                continue;
            if (path === NetworkService.path) {
                entries.push({
                    "path": path,
                    "label": I18n.tr("Network", "network locations page and sidebar entry"),
                    "icon": ""
                });
                continue;
            }
            const entry = BookmarksService.entries.find(candidate => candidate.path === path);
            if (entry)
                entries.push(entry);
        }
        return entries;
    }

    function savedEntries() {
        const mounts = mountedPaths();
        return BookmarksService.entries.filter(entry => {
            if (essentialPaths.includes(entry.path))
                return false;
            return !mounts.includes(normalize(entry.path));
        });
    }

    function reorderSaved(orderedPaths) {
        const visible = new Set(bookmarks.map(entry => entry.path));
        const queue = orderedPaths.slice();
        const ordered = BookmarksService.paths.map(path => visible.has(path) ? queue.shift() : path);
        BookmarksService.reorder(ordered.concat(queue));
    }

    function kindOf(path) {
        switch (path) {
        case NetworkService.path:
            return "network";
        case RecentService.path:
            return "recent";
        case TrashService.path:
            return "trash";
        default:
            return "bookmark";
        }
    }

    function targetOf(entry) {
        return entry.path === TrashService.path ? TrashService.filesPath : entry.path;
    }

    function placeFor(path) {
        return DirectoryService.userDirs.find(dir => dir.path === path) ?? null;
    }

    function labelFor(entry) {
        const place = placeFor(entry.path);
        if (place?.key === "home" && entry.label === place.name)
            return I18n.tr("Home", "sidebar entry for the user's home folder");
        return entry.label;
    }

    function iconFor(entry) {
        return entry.icon || placeFor(entry.path)?.iconName || "folder";
    }

    component BookmarkRow: OmarchyPlaceRow {
        id: row

        required property var modelData
        property int orderIndex: -1
        property bool reorderable: false

        readonly property string kind: section.kindOf(modelData.path)
        readonly property bool recent: kind === "recent"
        readonly property bool network: kind === "network"
        readonly property bool trash: kind === "trash"
        readonly property bool saved: kind === "bookmark"
        readonly property string target: section.targetOf(modelData)
        readonly property bool dragging: reorderable && reorderList.draggingIndex === orderIndex

        label: network ? I18n.tr("Network", "network locations page and sidebar entry") : recent ? I18n.tr("Recent", "recent files view header and sidebar entry") : trash ? I18n.tr("Trash", "trash view header and sidebar entry") : section.labelFor(modelData)
        glyph: network ? "lan" : recent ? "history" : ""
        iconName: network || recent ? "" : trash ? (TrashService.count > 0 ? "user-trash-full" : "user-trash") : section.iconFor(modelData)
        tooltipText: {
            if (recent)
                return RecentService.count < 0 ? "" : RecentService.count > 0 ? FileFormat.count(RecentService.count) : I18n.tr("No recent files", "recent files view placeholder when nothing was opened yet");
            if (trash)
                return TrashService.count > 0 ? FileFormat.count(TrashService.count) : I18n.tr("Trash is empty", "trash view placeholder and sidebar tooltip when nothing is trashed");
            return modelData.path;
        }
        selected: target === section.currentPath
        dropPath: recent || network ? "" : target
        springLoad: saved
        z: dragging ? 1 : 0
        opacity: dragging ? Theme.pendingOpacity : 1
        onClicked: section.navigateRequested(target)
        onMiddleClicked: section.tabRequested(target)
        onDropped: drop => section.dropRequested(target, drop)
        onSpringLoaded: section.navigateRequested(target)

        Behavior on y {
            enabled: reorderList.animateLayout && !row.dragging && Theme.animationsEnabled
            NumberAnimation {
                duration: Theme.expressiveDurations.expressiveDefaultSpatial
                easing.type: Easing.BezierSpline
                easing.bezierCurve: Theme.expressiveCurves.expressiveFastSpatial
            }
        }

        TapHandler {
            acceptedButtons: Qt.RightButton
            onTapped: point => {
                const mapped = row.mapToItem(section, point.position.x, point.position.y);
                switch (row.kind) {
                case "network":
                    return;
                case "recent":
                    section.recentMenuRequested(mapped.x, mapped.y);
                    return;
                case "trash":
                    section.trashMenuRequested(mapped.x, mapped.y);
                    return;
                default:
                    section.menuRequested(row.modelData.path, mapped.x, mapped.y);
                }
            }
        }

        DragHandler {
            target: null
            xAxis.enabled: false
            enabled: row.reorderable && reorderList.count > 1
            onGrabChanged: (transition, point) => {
                switch (transition) {
                case PointerDevice.GrabExclusive:
                    reorderList.begin(row.orderIndex, reorderList.mapFromItem(null, point.scenePressPosition.x, point.scenePressPosition.y));
                    reorderList.dragTo(reorderList.mapFromItem(null, point.scenePosition.x, point.scenePosition.y));
                    return;
                case PointerDevice.UngrabExclusive:
                    reorderList.finish();
                    return;
                case PointerDevice.CancelGrabExclusive:
                    reorderList.cancel();
                    return;
                }
            }
            onCentroidChanged: {
                if (!active || !row.dragging)
                    return;
                reorderList.dragTo(reorderList.mapFromItem(null, centroid.scenePosition.x, centroid.scenePosition.y));
            }
        }
    }

    Repeater {
        model: section.essentials

        BookmarkRow {}
    }

    OmarchySidebarDivider {
        visible: section.essentials.length > 0 && section.bookmarks.length > 0
        height: visible ? implicitHeight : 0
    }

    DReorderList {
        id: reorderList

        width: section.width
        model: section.bookmarks
        spacing: Theme.spacingXXS
        onReordered: indices => section.reorderSaved(indices.map(index => section.bookmarks[index].path))

        delegate: BookmarkRow {
            required property int index

            orderIndex: index
            reorderable: true
            width: reorderList.width
        }
    }
}
