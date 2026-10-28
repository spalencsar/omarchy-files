pragma ComponentBehavior: Bound

import QtQuick
import qs.Common
import qs.Services
import qs.DCommon.Widgets

Item {
    id: root

    readonly property var activeJobs: FileOpsService.jobs.filter(job => job.state !== "done" && job.state !== "failed")
    readonly property var primaryJob: activeJobs.length > 0 ? activeJobs[0] : null
    readonly property real fraction: {
        const progress = primaryJob?.progress;
        if (!progress)
            return -1;
        const total = progress.bytesTotal > 0 ? progress.bytesTotal : progress.filesTotal;
        const done = progress.bytesTotal > 0 ? progress.bytesDone : progress.filesDone;
        return total > 0 ? Math.max(0, Math.min(1, done / total)) : -1;
    }
    readonly property bool indeterminate: fraction < 0 || primaryJob?.progress?.estimating === true

    visible: activeJobs.length > 0
    implicitWidth: Math.min(FilesMetrics.popoverWidth * 0.62, 280)
    implicitHeight: Theme.fieldHeightLarge
    z: 900

    onVisibleChanged: {
        if (!visible)
            details.close();
    }

    Rectangle {
        id: compact

        anchors.fill: parent
        radius: Theme.cornerRadiusS
        color: hover.hovered ? Theme.surfaceContainerHighest : Theme.surfaceContainerHigh
        border.width: Theme.layerOutlineWidth
        border.color: Theme.outlineStrong

        Behavior on color {
            enabled: Theme.animationsEnabled
            ColorAnimation {
                duration: Theme.shortDuration
                easing.type: Theme.standardEasing
            }
        }

        Row {
            anchors.fill: parent
            anchors.leftMargin: Theme.spacingM
            anchors.rightMargin: Theme.spacingM
            spacing: Theme.spacingS

            Item {
                anchors.verticalCenter: parent.verticalCenter
                width: Theme.iconSizeMedium
                height: width

                Canvas {
                    id: ring

                    anchors.fill: parent
                    readonly property real progress: root.indeterminate ? 0.28 : root.fraction

                    onProgressChanged: requestPaint()
                    onPaint: {
                        const context = getContext("2d");
                        const center = width / 2;
                        const lineWidth = Math.max(2, Theme.outlineWidth);
                        const radius = Math.max(1, center - lineWidth);
                        context.reset();
                        context.lineWidth = lineWidth;
                        context.lineCap = "round";
                        context.strokeStyle = Theme.outlineStrong;
                        context.beginPath();
                        context.arc(center, center, radius, 0, Math.PI * 2);
                        context.stroke();
                        if (progress <= 0)
                            return;
                        context.strokeStyle = Theme.primary;
                        context.beginPath();
                        context.arc(center, center, radius, -Math.PI / 2, -Math.PI / 2 + Math.PI * 2 * progress);
                        context.stroke();
                    }

                    NumberAnimation on rotation {
                        running: root.visible && root.indeterminate
                        from: 0
                        to: 360
                        duration: 1100
                        loops: Animation.Infinite
                    }
                }
            }

            StyledText {
                anchors.verticalCenter: parent.verticalCenter
                width: parent.width - Theme.iconSizeMedium - parent.spacing
                text: root.primaryJob ? JobLabels.summary(root.primaryJob) : ""
                color: Theme.surfaceText
                font.pixelSize: Theme.fontSizeMedium
                elide: Text.ElideRight
                maximumLineCount: 1
            }
        }

        HoverHandler {
            id: hover
            cursorShape: Qt.PointingHandCursor
        }

        TapHandler {
            acceptedButtons: Qt.LeftButton
            onTapped: details.visible ? details.close() : details.open()
        }

        Accessible.role: Accessible.Button
        Accessible.name: root.primaryJob ? JobLabels.summary(root.primaryJob) : I18n.tr("File operations", "jobs popover header")
    }

    JobsPopover {
        id: details

        x: 0
        y: -height - Theme.spacingXS
    }
}
