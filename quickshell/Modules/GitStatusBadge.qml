pragma ComponentBehavior: Bound

import QtQuick
import QtQuick.Controls
import qs.Common
import qs.DCommon.Widgets

Rectangle {
    id: badge

    required property string code
    property bool aggregate: false
    property real badgeSize: 16

    readonly property string symbol: {
        switch (code) {
        case "added": return "+";
        case "untracked": return "?";
        case "deleted": return "−";
        case "renamed": return "R";
        case "conflict": return "!";
        default: return "M";
        }
    }
    readonly property color statusColor: {
        switch (code) {
        case "added": return Theme.success;
        case "untracked": return Theme.primary;
        case "deleted": return Theme.error;
        case "renamed": return Theme.secondary;
        case "conflict": return Theme.error;
        default: return Theme.warning;
        }
    }
    readonly property string statusLabel: {
        switch (code) {
        case "added": return I18n.tr("Added to Git", "tooltip for a file added to the Git index");
        case "untracked": return I18n.tr("Not tracked by Git", "tooltip for a file Git does not track yet");
        case "deleted": return I18n.tr("Deleted in Git", "tooltip for a file deleted from the Git worktree");
        case "renamed": return I18n.tr("Renamed in Git", "tooltip for a file renamed in Git");
        case "conflict": return I18n.tr("Git conflict", "tooltip for a file with an unresolved Git conflict");
        default: return I18n.tr("Modified in Git", "tooltip for a file modified in the Git worktree");
        }
    }

    width: badgeSize
    height: badgeSize
    radius: Theme.cornerRadiusXS
    color: Theme.cardSurface
    border.width: Theme.outlineWidth
    border.color: statusColor
    opacity: aggregate ? 0.82 : 1
    Accessible.role: Accessible.StaticText
    Accessible.name: statusLabel
    ToolTip.visible: hover.hovered
    ToolTip.text: statusLabel
    ToolTip.delay: 350

    HoverHandler {
        id: hover
    }

    StyledText {
        anchors.centerIn: parent
        text: badge.symbol
        color: badge.statusColor
        font.pixelSize: Math.max(9, badge.badgeSize - 6)
        font.weight: Theme.fontWeightMedium
    }
}
