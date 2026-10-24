pragma Singleton
pragma ComponentBehavior: Bound

import QtQuick
import Quickshell
import qs.Services

Singleton {
    id: root

    property var transport: DuxFilesService

    property bool loaded: false
    property bool sandboxed: false
    property string appId: ""
    property bool hostFilesystem: true
    property string hostOsRoot: ""
    property var portals: ({})
    property var features: ({})
    property var thumbnailers: []
    property var optionalDependencies: ({})

    // Unknown features count as available so nothing flickers out at startup.
    function available(name) {
        const feature = features[name];
        return feature ? feature.available === true : true;
    }

    function degraded(name) {
        const feature = features[name];
        return feature ? feature.degraded === true : false;
    }

    function reason(name) {
        const feature = features[name];
        return feature ? (feature.reason || "") : "";
    }

    function refresh() {
        transport.call("system.capabilities", null, result => {
            if (result.error)
                return;
            root._apply(result);
        });
    }

    function _apply(data) {
        sandboxed = data.sandboxed === true;
        appId = data.appId || "";
        hostFilesystem = data.hostFilesystem === true;
        hostOsRoot = data.hostOsRoot || "";
        portals = data.portals || {};
        features = data.features || {};
        thumbnailers = data.thumbnailers || [];
        optionalDependencies = data.optionalDependencies || {};
        loaded = true;
    }

    Component.onCompleted: {
        if (transport.connected)
            refresh();
    }

    onTransportChanged: {
        if (transport.connected)
            refresh();
    }

    Connections {
        target: root.transport

        function onConnectedChanged() {
            if (root.transport.connected)
                root.refresh();
        }
    }
}
