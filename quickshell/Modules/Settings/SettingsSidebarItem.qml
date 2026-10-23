import QtQuick
import qs.Common
import qs.DCommon.Widgets

Rectangle {
    id: root

    property string iconName: ""
    property string title: ""
    property string hint: ""
    property string accent: ""
    property bool active: false
    property bool highlighted: false
    property bool isFirstInGroup: true
    property bool isLastInGroup: true
    property int matchCount: -1

    signal clicked

    activeFocusOnTab: enabled
    Accessible.role: Accessible.Button
    Accessible.name: title
    Accessible.description: matchCount > 0 ? I18n.tr("%1, %2 matches", "accessible description of a settings page while searching, name then count").arg(hint).arg(matchCount) : hint
    Accessible.onPressAction: clicked()

    Keys.onPressed: event => {
        switch (event.key) {
        case Qt.Key_Space:
        case Qt.Key_Return:
        case Qt.Key_Enter:
            root.clicked();
            event.accepted = true;
            break;
        }
    }

    readonly property color glyphColor: root.active ? Theme.primary : Theme.surfaceVariantText
    readonly property real topRadius: isFirstInGroup ? Theme.groupedListOuterRadius : Theme.groupedListInnerRadius
    readonly property real bottomRadius: isLastInGroup ? Theme.groupedListOuterRadius : Theme.groupedListInnerRadius

    width: parent?.width ?? 0
    height: Math.max(SettingsMetrics.navItemMinHeight, textColumn.implicitHeight + Theme.spacingS * 2)
    topLeftRadius: topRadius
    topRightRadius: topRadius
    bottomLeftRadius: bottomRadius
    bottomRightRadius: bottomRadius
    border.width: active ? Theme.outlineWidth : 0
    border.color: active ? Theme.outlineVariant : "transparent"
    color: {
        if (active)
            return Theme.surfaceContainerLow;
        if (highlighted)
            return Theme.blend(SettingsMetrics.rowColor, Theme.primary, SettingsMetrics.highlightBlend);
        return SettingsMetrics.rowColor;
    }

    Behavior on color {
        enabled: Theme.animationsEnabled
        ColorAnimation {
            duration: SettingsMetrics.fadeDuration
            easing.type: Easing.BezierSpline
            easing.bezierCurve: Theme.expressiveCurves.expressiveEffects
        }
    }

    Rectangle {
        anchors.fill: parent
        topLeftRadius: root.topRadius
        topRightRadius: root.topRadius
        bottomLeftRadius: root.bottomRadius
        bottomRightRadius: root.bottomRadius
        color: Theme.surfaceText
        opacity: mouseArea.pressed ? Theme.stateLayerPressed : (mouseArea.containsMouse ? Theme.stateLayerHover : 0)

        Behavior on opacity {
            enabled: Theme.animationsEnabled
            NumberAnimation {
                duration: SettingsMetrics.fadeDuration
                easing.type: Easing.BezierSpline
                easing.bezierCurve: Theme.expressiveCurves.expressiveEffects
            }
        }
    }

    DRipple {
        id: ripple
        rippleColor: Theme.surfaceText
        topLeftRadius: root.topRadius
        topRightRadius: root.topRadius
        bottomLeftRadius: root.bottomRadius
        bottomRightRadius: root.bottomRadius
    }

    Rectangle {
        anchors.fill: parent
        anchors.margins: Theme.focusRingWidth / 2
        topLeftRadius: Math.max(0, root.topRadius - Theme.focusRingWidth / 2)
        topRightRadius: topLeftRadius
        bottomLeftRadius: Math.max(0, root.bottomRadius - Theme.focusRingWidth / 2)
        bottomRightRadius: bottomLeftRadius
        color: "transparent"
        border.width: Theme.focusRingWidth
        border.color: Theme.focusRingColor
        visible: root.activeFocus
    }

    Item {
        id: iconSlot
        width: SettingsMetrics.navIconSize
        height: SettingsMetrics.navIconSize
        anchors.left: parent.left
        anchors.leftMargin: Theme.spacingL
        anchors.verticalCenter: parent.verticalCenter

        DIcon {
            anchors.centerIn: parent
            name: root.iconName
            size: Theme.iconSizeMedium
            color: root.glyphColor
        }
    }

    StyledText {
        id: countBadge
        anchors.right: parent.right
        anchors.rightMargin: Theme.spacingL
        anchors.verticalCenter: parent.verticalCenter
        visible: root.matchCount > 0
        text: root.matchCount
        font.pixelSize: Theme.fontSizeSmall
        font.weight: Theme.fontWeightMedium
        color: Theme.primary
    }

    Column {
        id: textColumn
        anchors.left: iconSlot.right
        anchors.leftMargin: Theme.spacingM
        anchors.right: countBadge.visible ? countBadge.left : parent.right
        anchors.rightMargin: countBadge.visible ? Theme.spacingM : Theme.spacingL
        anchors.verticalCenter: parent.verticalCenter
        spacing: Theme.spacingXXS

        StyledText {
            width: parent.width
            text: root.title
            font.pixelSize: Theme.fontSizeMedium
            font.weight: Theme.fontWeightMedium
            color: Theme.surfaceText
            elide: Text.ElideRight
            horizontalAlignment: Text.AlignLeft
        }

        StyledText {
            width: parent.width
            text: root.hint
            font.pixelSize: Theme.fontSizeSmall
            color: Theme.surfaceVariantText
            wrapMode: Text.WordWrap
            maximumLineCount: 2
            elide: Text.ElideRight
            visible: root.hint !== ""
            horizontalAlignment: Text.AlignLeft
        }
    }

    MouseArea {
        id: mouseArea
        anchors.fill: parent
        hoverEnabled: true
        cursorShape: Qt.PointingHandCursor
        onPressed: mouse => ripple.trigger(mouse.x, mouse.y)
        onClicked: root.clicked()
    }
}
