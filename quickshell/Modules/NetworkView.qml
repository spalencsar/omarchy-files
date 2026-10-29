pragma ComponentBehavior: Bound

import QtQuick
import QtQuick.Layouts
import qs.Common
import qs.Services
import qs.DCommon.Widgets

Rectangle {
    id: root

    property bool outlined: false
    property string errorText: ""
    property string busyId: ""
    property string pendingUri: ""
    property string selectedScheme: ""

    readonly property var recentServers: NetworkService.servers.filter(server => !NetworkService.networkMounts.some(mount => root.sameServer(mount.uri, server.uri)))
    readonly property int connectionControlHeight: Theme.buttonHeightXS
    readonly property var protocols: NetworkService.schemes.map(scheme => ({
            "scheme": scheme,
            "label": protocolLabel(scheme)
        }))

    signal navigateRequested(string path)
    signal tabRequested(string path)
    signal focusRequested
    signal backRequested
    signal forwardRequested

    function sameServer(left, right) {
        const normalize = uri => (uri ?? "").replace(/\/+$/, "");
        return normalize(left) === normalize(right);
    }

    function localPath(result) {
        if (result.error)
            return "";
        return result.location || result.mount?.path || "";
    }

    function protocolLabel(scheme) {
        switch (scheme) {
        case "sftp": return I18n.tr("SFTP", "network protocol name");
        case "ssh": return I18n.tr("SSH", "network protocol name");
        case "smb": return I18n.tr("SMB / Windows share", "network protocol name");
        case "nfs": return I18n.tr("NFS", "network protocol name");
        case "dav": return I18n.tr("WebDAV", "network protocol name");
        case "davs": return I18n.tr("WebDAV (secure)", "network protocol name");
        case "ftp": return I18n.tr("FTP", "network protocol name");
        case "ftps": return I18n.tr("FTPS", "network protocol name");
        case "afp": return I18n.tr("AFP", "network protocol name");
        default: return scheme.toUpperCase();
        }
    }

    function schemeForLabel(label) {
        return protocols.find(protocol => protocol.label === label)?.scheme ?? "";
    }

    function typedScheme(value) {
        return value.trim().match(/^([A-Za-z][A-Za-z0-9+.-]*):\/\//)?.[1]?.toLowerCase() ?? "";
    }

    function placeholder() {
        switch (selectedScheme) {
        case "sftp": return "user@server/path";
        case "ssh": return "user@server/path";
        case "smb": return "server/share";
        case "nfs": return "server/export";
        case "dav":
        case "davs": return "server/path";
        case "ftp":
        case "ftps": return "user@server/path";
        case "afp": return "server/share";
        default: return I18n.tr("Server address", "placeholder in the network page connection field");
        }
    }

    function ensureScheme() {
        if (protocols.some(protocol => protocol.scheme === selectedScheme))
            return;
        selectedScheme = protocols[0]?.scheme ?? "";
    }

    function chooseProtocol(label) {
        const scheme = schemeForLabel(label);
        if (scheme === "" || scheme === selectedScheme)
            return;
        selectedScheme = scheme;
        if (typedScheme(address.text) !== "")
            address.text = address.text.replace(/^[A-Za-z][A-Za-z0-9+.-]*:\/\//, scheme + "://");
    }

    function connectionTarget(value) {
        const target = value.trim();
        if (target === "" || typedScheme(target) !== "")
            return target;
        return selectedScheme === "" ? target : selectedScheme + "://" + target.replace(/^\/+/, "");
    }

    function connect(uri) {
        const target = connectionTarget(uri || address.text);
        if (target === "" || NetworkService.connecting)
            return;
        errorText = "";
        pendingUri = target;
        NetworkService.connectTo(target, result => {
            pendingUri = "";
            const path = localPath(result);
            if (path !== "") {
                address.text = "";
                root.navigateRequested(path);
                return;
            }
            errorText = result.error || I18n.tr("Connected, but the share has no local folder. The gvfs FUSE service may not be running.", "network page error when a mounted share has no fuse path");
            address.forceActiveFocus();
        });
    }

    function disconnect(mount) {
        if (!mount?.id || busyId !== "")
            return;
        busyId = mount.id;
        NetworkService.disconnect(mount.id, () => busyId = "");
    }

    function focusBody() {
        address.forceActiveFocus(Qt.TabFocusReason);
    }

    function refresh() {
        NetworkService.refresh();
    }

    radius: FilesMetrics.paneRadius
    color: Theme.foregroundColor(Theme.cardSurface, true)
    border.width: outlined ? Theme.outlineWidthFocused : 0
    border.color: Theme.primary

    DFlickable {
        id: scroll

        anchors.fill: parent
        contentHeight: content.implicitHeight + FilesMetrics.paneBodyMargin * 2
        clip: true

        Column {
            id: content

            x: FilesMetrics.paneBodyMarginH
            y: FilesMetrics.paneBodyMargin
            width: scroll.width - FilesMetrics.paneBodyMarginH * 2
            spacing: Theme.spacingL

            StyledText {
                width: parent.width
                text: I18n.tr("Network", "network locations page and sidebar entry")
                color: Theme.surfaceText
                font.pixelSize: Theme.fontSizeLarge
                font.weight: Theme.fontWeightMedium
            }

            RowLayout {
                width: parent.width
                spacing: Theme.spacingS

                DDropdown {
                    Layout.alignment: Qt.AlignTop
                    Layout.preferredWidth: dropdownWidth
                    visible: root.protocols.length > 0
                    showLabel: true
                    dropdownWidth: Math.min(Theme.spacingXL * 6, content.width * 0.28)
                    triggerHeight: root.connectionControlHeight
                    options: root.protocols.map(protocol => protocol.label)
                    currentValue: root.protocolLabel(root.selectedScheme)
                    onValueChanged: value => root.chooseProtocol(value)
                }

                DTextField {
                    id: address

                    Layout.fillWidth: true
                    outlined: true
                    controlHeight: root.connectionControlHeight
                    cornerRadius: Theme.cornerRadiusXS
                    isError: root.errorText !== ""
                    supportingText: root.errorText
                    placeholderText: root.placeholder()
                    enabled: NetworkService.available && !NetworkService.connecting
                    onTextChanged: {
                        if (root.errorText !== "")
                            root.errorText = "";
                        const scheme = root.typedScheme(text);
                        if (root.protocols.some(protocol => protocol.scheme === scheme))
                            root.selectedScheme = scheme;
                    }
                    onAccepted: root.connect("")
                }

                DButton {
                    Layout.alignment: Qt.AlignTop
                    text: I18n.tr("Connect", "connect to server action")
                    buttonHeight: root.connectionControlHeight
                    shape: "square"
                    backgroundColor: Theme.primary
                    textColor: Theme.primaryText
                    enabled: NetworkService.available && !NetworkService.connecting && address.text.trim() !== "" && (root.typedScheme(address.text) !== "" || root.selectedScheme !== "")
                    onClicked: root.connect("")
                }

                DActionButton {
                    Layout.alignment: Qt.AlignTop
                    buttonSize: root.connectionControlHeight
                    iconName: "info"
                    iconColor: Theme.surfaceVariantText
                    backgroundColor: "transparent"
                    tooltipText: I18n.tr("Available protocols: %1. You can also paste a complete address such as sftp://host/path.", "network page help listing supported server protocols, %1 is the protocol list").arg(root.protocols.map(protocol => protocol.label).join(", "))
                }
            }

            StyledText {
                width: parent.width
                visible: !NetworkService.available
                text: NetworkService.reason || I18n.tr("Network locations are unavailable because gvfs is not running", "network page explanation when the network service is missing")
                color: Theme.error
                font.pixelSize: Theme.fontSizeSmall
                wrapMode: Text.Wrap
            }

            Column {
                width: parent.width
                spacing: Theme.spacingXS

                StyledText {
                    width: parent.width
                    text: I18n.tr("Connected", "heading above active network shares")
                    color: Theme.surfaceText
                    font.pixelSize: Theme.fontSizeMedium
                    font.weight: Theme.fontWeightMedium
                }

                StyledText {
                    width: parent.width
                    visible: NetworkService.networkMounts.length === 0
                    text: I18n.tr("No network locations are connected", "empty state on the network locations page")
                    color: Theme.surfaceVariantText
                    font.pixelSize: Theme.fontSizeSmall
                    leftPadding: Theme.spacingS
                    topPadding: Theme.spacingS
                    bottomPadding: Theme.spacingS
                }

                Repeater {
                    model: NetworkService.networkMounts

                    OmarchyPlaceRow {
                        required property var modelData

                        label: modelData.name || modelData.uri
                        subtitle: modelData.uri || ""
                        tooltipText: modelData.path || modelData.uri || ""
                        iconName: modelData.iconName || "folder-remote"
                        actionIcon: "eject"
                        actionTooltip: I18n.tr("Disconnect", "network mount action")
                        busy: root.busyId === modelData.id
                        selected: false
                        onClicked: root.navigateRequested(modelData.path)
                        onMiddleClicked: root.tabRequested(modelData.path)
                        onActionTriggered: root.disconnect(modelData)
                    }
                }
            }

            Column {
                width: parent.width
                visible: root.recentServers.length > 0
                spacing: Theme.spacingXS

                StyledText {
                    width: parent.width
                    text: I18n.tr("Previous", "heading above previously used servers")
                    color: Theme.surfaceText
                    font.pixelSize: Theme.fontSizeMedium
                    font.weight: Theme.fontWeightMedium
                }

                Repeater {
                    model: root.recentServers

                    OmarchyPlaceRow {
                        required property var modelData

                        label: modelData.name || modelData.uri
                        subtitle: modelData.name ? modelData.uri : ""
                        tooltipText: modelData.uri
                        iconName: "folder-remote"
                        actionIcon: "close"
                        actionTooltip: I18n.tr("Forget", "action removing a server from network history")
                        busy: NetworkService.connecting && root.pendingUri === modelData.uri
                        onClicked: root.connect(modelData.uri)
                        onActionTriggered: NetworkService.forget(modelData.uri)
                    }
                }
            }
        }
    }

    TapHandler {
        acceptedButtons: Qt.BackButton | Qt.ForwardButton
        onTapped: (point, button) => {
            root.focusRequested();
            if (button === Qt.BackButton) {
                root.backRequested();
                return;
            }
            root.forwardRequested();
        }
    }

    Connections {
        target: NetworkService

        function onSchemesChanged() {
            root.ensureScheme();
        }
    }

    Component.onCompleted: ensureScheme()
}
