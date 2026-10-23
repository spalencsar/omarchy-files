pragma ComponentBehavior: Bound

import QtQuick
import qs.Common
import qs.DCommon.Common
import qs.DCommon.FileBrowser
import qs.DCommon.Widgets

Row {
    id: control

    property string sortKey: "name"
    property bool descending: false
    property bool showLabel: false
    property var keys: defaultKeys
    readonly property var defaultKeys: [
        { "key": "name", "icon": "sort_by_alpha", "label": I18n.tr("Name", "file browser sort criterion option") },
        { "key": "size", "icon": "straighten", "label": I18n.tr("Size", "file browser sort criterion option") },
        { "key": "mtime", "icon": "history", "label": I18n.tr("Modified", "file browser sort criterion option") },
        { "key": "type", "icon": "category", "label": I18n.tr("Type", "file browser sort criterion option") }
    ]
    readonly property var current: keys.find(entry => entry.key === sortKey) ?? keys[0]

    signal sortKeySelected(string key)
    signal directionToggled

    spacing: Theme.spacingXS

    StyledButton {
        id: keyButton
        width: control.showLabel ? Math.max(FilesMetrics.topbarControlSize, labelRow.implicitWidth + Theme.spacingL * 2) : FilesMetrics.topbarControlSize
        height: FilesMetrics.topbarControlSize
        radius: Theme.cornerRadiusS
        color: Theme.surfaceContainerLowest
        border.width: Theme.outlineWidth
        border.color: Theme.outlineVariant
        Accessible.name: I18n.tr("Sort By", "file browser sort menu section header")
        onClicked: {
            menu.currentValue = control.current.label;
            menu.openDropdownMenu();
        }

        StateLayer { control: keyButton; stateColor: Theme.surfaceText }

        Row {
            id: labelRow
            anchors.centerIn: parent
            spacing: Theme.spacingS
            DIcon { anchors.verticalCenter: parent.verticalCenter; name: control.current.icon; size: Theme.iconSizeMedium; color: Theme.surfaceText }
            StyledText {
                anchors.verticalCenter: parent.verticalCenter
                visible: control.showLabel
                text: control.current.label
                color: Theme.surfaceText
                font.pixelSize: Theme.fontSizeMedium
                font.weight: Theme.fontWeightMedium
            }
        }

        DDropdown {
            id: menu
            showTrigger: false
            popupAnchorItem: keyButton
            focusReturnTarget: keyButton
            alignPopupRight: !I18n.isRtl
            popupWidth: FileBrowserMetrics.menuWidth
            options: control.keys.map(entry => entry.label)
            optionIcons: control.keys.map(entry => entry.icon)
            onValueChanged: value => {
                const index = options.indexOf(value);
                if (index >= 0)
                    control.sortKeySelected(control.keys[index].key);
            }
        }
    }

    FlatButton {
        iconName: control.descending ? "arrow_downward" : "arrow_upward"
        tooltipText: control.descending ? I18n.tr("Descending", "file browser sort order option") : I18n.tr("Ascending", "file browser sort order option")
        onClicked: control.directionToggled()
    }
}
