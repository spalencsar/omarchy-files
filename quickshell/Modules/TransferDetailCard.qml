pragma ComponentBehavior: Bound

import QtQuick
import QtQuick.Controls
import QtQuick.Layouts
import qs.Common
import qs.DCommon.Common as DC
import qs.DCommon.Widgets

Popup {
    id: root

    required property var job
    readonly property bool hovered: hover.hovered
    readonly property real fraction: {
        const progress = job?.progress;
        if (!progress)
            return 0;
        const total = progress.bytesTotal > 0 ? progress.bytesTotal : progress.filesTotal;
        const done = progress.bytesTotal > 0 ? progress.bytesDone : progress.filesDone;
        return total > 0 ? Math.max(0, Math.min(1, done / total)) : 0;
    }

    width: FilesMetrics.popoverWidth
    margins: 0
    padding: Theme.spacingL
    modal: false
    focus: false
    closePolicy: Popup.NoAutoClose

    background: Item {
        DC.ElevationShadow {
            anchors.fill: parent
            level: Theme.elevationLevel2
            fallbackOffset: Theme.spacingXS
            targetRadius: Theme.cornerRadiusL
            targetColor: Theme.surfaceContainer
            borderColor: Theme.outlineMedium
            borderWidth: Theme.layerOutlineWidth
            shadowEnabled: Theme.elevationEnabled && SettingsData.popoutElevationEnabled
        }
    }

    contentItem: ColumnLayout {
        spacing: Theme.spacingS

        StyledText {
            Layout.fillWidth: true
            text: JobLabels.summary(root.job)
            color: Theme.surfaceText
            font.pixelSize: Theme.fontSizeMedium
            font.weight: Theme.fontWeightMedium
            elide: Text.ElideMiddle
            maximumLineCount: 1
        }

        Rectangle {
            Layout.fillWidth: true
            Layout.preferredHeight: Theme.spacingXS
            radius: Theme.cornerRadiusFull
            color: Theme.surfaceContainerHigh

            Rectangle {
                width: parent.width * root.fraction
                height: parent.height
                radius: parent.radius
                color: Theme.primary

                Behavior on width {
                    enabled: Theme.animationsEnabled
                    NumberAnimation {
                        duration: Theme.shortDuration
                        easing.type: Theme.standardEasing
                    }
                }
            }
        }

        RowLayout {
            Layout.fillWidth: true
            spacing: Theme.spacingS

            StyledText {
                Layout.fillWidth: true
                text: JobLabels.currentItem(root.job)
                color: Theme.surfaceVariantText
                font.pixelSize: Theme.fontSizeSmall
                elide: Text.ElideMiddle
                maximumLineCount: 1
            }

            StyledText {
                text: JobLabels.progress(root.job)
                color: Theme.surfaceVariantText
                font.pixelSize: Theme.fontSizeSmall
                font.features: {
                    "tnum": 1
                }
            }
        }

        StyledText {
            Layout.fillWidth: true
            visible: text !== ""
            text: JobLabels.route(root.job)
            color: Theme.surfaceVariantText
            font.pixelSize: Theme.fontSizeSmall
            elide: Text.ElideMiddle
            maximumLineCount: 1
        }

        StyledText {
            Layout.fillWidth: true
            visible: text !== ""
            text: FilesFormat.transferHint(root.job?.progress)
            color: Theme.primary
            font.pixelSize: Theme.fontSizeSmall
            font.features: {
                "tnum": 1
            }
        }

        HoverHandler {
            id: hover
        }
    }
}
