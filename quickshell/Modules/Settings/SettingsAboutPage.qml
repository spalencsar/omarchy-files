pragma ComponentBehavior: Bound

import QtQuick
import qs.Common
import qs.Services
import qs.DCommon.Widgets

Item {
    id: root

    LayoutMirroring.enabled: I18n.isRtl
    LayoutMirroring.childrenInherit: true

    readonly property string githubUrl: "https://github.com/spalencsar/omarchy-files"
    readonly property string docsUrl: githubUrl + "#readme"
    readonly property string licenseUrl: githubUrl + "/blob/master/LICENSE"
    readonly property string version: DuxFilesService.daemonVersion.match(/^v?(\d+\.\d+)/)?.[1] ?? ""

    readonly property var featureRows: [
        {
            "id": "hostFilesystem",
            "label": I18n.tr("Host filesystem", "capability row on the about page")
        },
        {
            "id": "trash",
            "label": I18n.tr("Trash", "capability row on the about page")
        },
        {
            "id": "thumbnails",
            "label": I18n.tr("Thumbnails", "capability row on the about page")
        },
        {
            "id": "devices",
            "label": I18n.tr("Devices", "sidebar section listing drives and volumes")
        },
        {
            "id": "network",
            "label": I18n.tr("Network", "sidebar section listing network locations")
        },
        {
            "id": "terminal",
            "label": I18n.tr("Open in terminal", "context menu entry opening a terminal")
        },
        {
            "id": "run",
            "label": I18n.tr("Run programs", "capability row on the about page")
        },
        {
            "id": "openWith",
            "label": I18n.tr("Open with", "open with dialog header")
        },
        {
            "id": "setDefault",
            "label": I18n.tr("Default applications", "capability row on the about page")
        },
        {
            "id": "elevation",
            "label": I18n.tr("Retry as administrator", "file operation error dialog action")
        }
    ]

    readonly property var optionalFeatureRows: [
        {
            "id": "git",
            "label": I18n.tr("Git integration", "optional feature row on the about page"),
            "availableSubtitle": I18n.tr("Repository status and staging actions are available.", "optional feature status on the about page"),
            "install": "sudo pacman -S git"
        },
        {
            "id": "dsearch",
            "label": I18n.tr("Indexed search", "optional feature row on the about page"),
            "availableSubtitle": I18n.tr("DankSearch is searching from its index.", "optional feature status on the about page"),
            "installedSubtitle": I18n.tr("DankSearch is installed, but its user service is not running.", "optional feature status on the about page"),
            "install": "yay -S dsearch-bin && systemctl --user enable --now dsearch"
        },
        {
            "id": "bubblewrap",
            "label": I18n.tr("Sandboxed thumbnails", "optional feature row on the about page"),
            "availableSubtitle": I18n.tr("External thumbnailers run in a sandbox.", "optional feature status on the about page"),
            "install": "sudo pacman -S bubblewrap"
        },
        {
            "id": "ffmpegthumbnailer",
            "label": I18n.tr("Video thumbnails", "optional feature row on the about page"),
            "availableSubtitle": I18n.tr("Video thumbnails are available.", "optional feature status on the about page"),
            "install": "sudo pacman -S ffmpegthumbnailer"
        },
        {
            "id": "poppler",
            "label": I18n.tr("PDF thumbnails", "optional feature row on the about page"),
            "availableSubtitle": I18n.tr("PDF thumbnails are available.", "optional feature status on the about page"),
            "install": "sudo pacman -S poppler"
        },
        {
            "id": "pdfPreview",
            "label": I18n.tr("PDF preview", "optional feature row on the about page"),
            "kind": "pdfPreview",
            "availableSubtitle": I18n.tr("PDF files can be previewed in the file manager.", "optional feature status on the about page"),
            "install": "sudo pacman -S qt6-webengine"
        },
        {
            "id": "network",
            "label": I18n.tr("Network locations", "optional feature row on the about page"),
            "kind": "network",
            "availableSubtitle": I18n.tr("Network shares can be connected and browsed.", "optional feature status on the about page"),
            "install": "sudo pacman -S gvfs"
        },
        {
            "id": "devices",
            "label": I18n.tr("Drives and removable media", "optional feature row on the about page"),
            "kind": "devices",
            "availableSubtitle": I18n.tr("Drives can be mounted and ejected.", "optional feature status on the about page"),
            "install": "sudo pacman -S udisks2"
        }
    ]

    function host(url) {
        return url.replace(/^https?:\/\//, "").replace(/\/$/, "");
    }

    function statusFor(id) {
        const feature = SandboxService.features[id];
        if (!feature)
            return {
                "text": I18n.tr("Unknown", "capability status when the daemon did not report it"),
                "color": Theme.surfaceVariantText
            };
        if (!feature.available)
            return {
                "text": FilesFormat.featureUnavailable(feature.reason),
                "color": Theme.error
            };
        if (feature.degraded)
            return {
                "text": FilesFormat.featureUnavailable(feature.reason),
                "color": Theme.tertiary
            };
        return {
            "text": I18n.tr("Available", "capability status on the about page"),
            "color": Theme.success
        };
    }

    function optionalStatus(row) {
        const dependency = SandboxService.optionalDependencies[row.id] || {};
        if (row.id === "dsearch") {
            if (!dependency.installed)
                return optionalMissing(row);
            if (dependency.active)
                return optionalAvailable(row, I18n.tr("Active", "optional search index status on the about page"));
            return {
                "text": I18n.tr("Installed", "optional package status on the about page"),
                "color": Theme.tertiary,
                "subtitle": row.installedSubtitle
            };
        }
        if (row.kind === "network" || row.kind === "devices") {
            const available = row.kind === "network" ? NetworkService.available : DevicesService.available;
            if (available)
                return optionalAvailable(row, I18n.tr("Available", "optional feature status on the about page"));
            return {
                "text": I18n.tr("Unavailable", "optional runtime feature status on the about page"),
                "color": Theme.error,
                "subtitle": row.kind === "network" ? (NetworkService.reason || I18n.tr("Install: %1", "optional feature install command on the about page").arg(row.install)) : (DevicesService.reason || I18n.tr("Install: %1", "optional feature install command on the about page").arg(row.install))
            };
        }
        if (row.kind === "pdfPreview") {
            if (PdfService.available)
                return optionalAvailable(row, I18n.tr("Available", "optional feature status on the about page"));
            return optionalMissing(row);
        }
        if (dependency.installed)
            return optionalAvailable(row, I18n.tr("Installed", "optional package status on the about page"));
        return optionalMissing(row);
    }

    function optionalAvailable(row, text) {
        return {
            "text": text,
            "color": Theme.success,
            "subtitle": row.availableSubtitle
        };
    }

    function optionalMissing(row) {
        return {
            "text": I18n.tr("Not installed", "optional feature status on the about page"),
            "color": Theme.surfaceVariantText,
            "subtitle": I18n.tr("This optional component is not installed.", "optional feature status on the about page"),
            "missing": true
        };
    }

    Component.onCompleted: PdfService.ensureProbed()

    SettingsPage {
        OmarchyAboutHeader { version: root.version ? `omarchy-files ${root.version}` : "omarchy-files" }

        SettingsGroup {
            SettingsLinkRow {
                iconName: "menu_book"
                title: I18n.tr("Docs", "about page resource button")
                keywords: I18n.tr("documentation help manual", "settings search keywords, space separated, add synonyms in your language")
                subtitle: root.host(root.docsUrl)
                url: root.docsUrl
            }
        }

        SettingsCard {
            title: I18n.tr("Backend", "about page daemon section header")

            SettingsRow {
                title: I18n.tr("Version", "about page daemon version column label")
                keywords: I18n.tr("daemon release build", "settings search keywords, space separated, add synonyms in your language")
                trailingBadge: DuxFilesService.daemonVersion || "—"
            }

            SettingsRow {
                title: I18n.tr("API", "about page daemon api version column label")
                keywords: I18n.tr("daemon protocol", "settings search keywords, space separated, add synonyms in your language")
                trailingBadge: DuxFilesService.apiVersion > 0 ? `v${DuxFilesService.apiVersion}` : "—"
            }

            SettingsRow {
                title: I18n.tr("Status", "about page daemon status column label")
                keywords: I18n.tr("daemon connected offline socket", "settings search keywords, space separated, add synonyms in your language")
                trailingBadge: DuxFilesService.connected ? I18n.tr("Connected", "about page daemon status value") : I18n.tr("Offline", "about page daemon status value")

                DBadge {
                    color: DuxFilesService.connected ? Theme.success : Theme.error
                }
            }

            SettingsRow {
                title: I18n.tr("Socket", "about page daemon socket row label")
                keywords: I18n.tr("daemon path unix", "settings search keywords, space separated, add synonyms in your language")
                subtitle: DuxFilesService.socketPath || "—"
            }

            SettingsRow {
                visible: SandboxService.sandboxed
                title: I18n.tr("Sandbox", "about page row naming the flatpak id")
                keywords: I18n.tr("flatpak app id", "settings search keywords, space separated, add synonyms in your language")
                subtitle: SandboxService.appId
            }
        }

        SettingsCard {
            title: I18n.tr("Capabilities", "about page daemon capabilities label")

            Repeater {
                model: root.featureRows

                SettingsRow {
                    required property var modelData

                    readonly property var status: root.statusFor(modelData.id)

                    title: modelData.label
                    keywords: I18n.tr("capability available sandbox", "settings search keywords, space separated, add synonyms in your language")
                    trailingBadge: status.text
                    trailingBadgeColor: status.color
                }
            }
        }

        SettingsCard {
            title: I18n.tr("Optional features", "about page section listing optional packages")

            Repeater {
                model: root.optionalFeatureRows

                SettingsRow {
                    required property var modelData

                    readonly property var status: root.optionalStatus(modelData)

                    title: modelData.label
                    subtitle: status.subtitle
                    keywords: I18n.tr("optional packages install git repository dsearch thumbnails preview network drives", "settings search keywords, space separated, add synonyms in your language")
                    trailingBadge: status.text
                    trailingBadgeColor: status.color

                    body: Rectangle {
                        visible: status.missing === true
                        width: parent.width
                        implicitHeight: helpContent.implicitHeight + Theme.spacingL * 2
                        radius: Theme.cornerRadiusS
                        color: Theme.surfaceContainerLow
                        border.width: Theme.outlineWidth
                        border.color: Theme.outlineVariant

                        Column {
                            id: helpContent

                            anchors.fill: parent
                            anchors.margins: Theme.spacingL
                            spacing: Theme.spacingS

                            StyledText {
                                width: parent.width
                                text: I18n.tr("Install in a terminal", "optional feature help box heading")
                                font.pixelSize: Theme.fontSizeSmall
                                font.weight: Theme.fontWeightMedium
                                color: Theme.surfaceText
                            }

                            StyledText {
                                width: parent.width
                                text: modelData.install
                                font.pixelSize: Theme.fontSizeSmall
                                font.family: "monospace"
                                color: Theme.primary
                                wrapMode: Text.WordWrap
                            }

                            Row {
                                spacing: Theme.spacingS

                                DButton {
                                    text: I18n.tr("Copy command", "optional feature help box action")
                                    buttonHeight: Theme.buttonHeightXS
                                    backgroundColor: "transparent"
                                    textColor: Theme.primary
                                    onClicked: Paths.copyPathToClipboard(modelData.install)
                                }

                                StyledText {
                                    anchors.verticalCenter: parent.verticalCenter
                                    text: I18n.tr("Restart Omarchy Files afterwards to refresh this status.", "optional feature help box restart note")
                                    font.pixelSize: Theme.fontSizeSmall
                                    color: Theme.surfaceVariantText
                                }
                            }
                        }
                    }
                }
            }
        }

        StyledText {
            anchors.horizontalCenter: parent.horizontalCenter
            text: I18n.tr("MIT License", "about page license footer link")
            font.pixelSize: Theme.fontSizeMedium
            color: Theme.primary

            MouseArea {
                anchors.fill: parent
                cursorShape: Qt.PointingHandCursor
                onClicked: DuxFilesService.openUri(root.licenseUrl)
            }
        }
    }
}
