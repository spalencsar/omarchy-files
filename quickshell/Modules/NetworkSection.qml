pragma ComponentBehavior: Bound

import QtQuick
import qs.Common
import qs.Services
import qs.DCommon.Widgets
import qs.DCommon.FileBrowser

OmarchyPlaceSection {
    id: section

    property string busyId: ""
    property string currentPath: ""

    signal navigateRequested(string path)
    signal tabRequested(string path)
    signal dropRequested(string path, var drop)
    signal connectRequested

    emptyText: {
        if (NetworkService.available)
            return "";
        if (!SandboxService.loaded)
            return "";
        if (!SandboxService.available("network"))
            return FilesFormat.featureUnavailable(SandboxService.reason("network"));
        return I18n.tr("Network locations are unavailable because gvfs is not running", "sidebar explanation when the network service is missing");
    }
    showEmpty: true

    Repeater {
        model: NetworkService.networkMounts

        OmarchyPlaceRow {
            required property var modelData

            label: modelData.name
            iconName: modelData.iconName
            actionIcon: "eject"
            actionTooltip: I18n.tr("Disconnect", "network mount action")
            busy: section.busyId === modelData.id
            dropPath: modelData.path
            selected: modelData.path === section.currentPath
            onClicked: section.navigateRequested(modelData.path)
            onMiddleClicked: section.tabRequested(modelData.path)
            onDropped: drop => section.dropRequested(modelData.path, drop)
            onSpringLoaded: section.navigateRequested(modelData.path)
            onActionTriggered: {
                section.busyId = modelData.id;
                NetworkService.disconnect(modelData.id, () => section.busyId = "");
            }
        }
    }

    OmarchyPlaceRow {
        visible: NetworkService.available
        label: I18n.tr("Connect to Server…", "sidebar entry opening the connect to server dialog")
        iconName: "folder-remote"
        busy: NetworkService.connecting
        onClicked: section.connectRequested()
    }
}
