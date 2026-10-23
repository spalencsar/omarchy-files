import QtQuick
import qs.Common
import qs.DCommon.Widgets

SettingsRow {
    id: root

    property string text: ""
    property string description: ""
    property string currentValue: ""
    property alias options: dropdown.options
    property alias enableFuzzySearch: dropdown.enableFuzzySearch
    property alias maxPopupHeight: dropdown.maxPopupHeight
    property alias dropdownWidth: dropdown.dropdownWidth

    signal valueChanged(string value)

    title: text
    subtitle: description
    onCurrentValueChanged: dropdown.currentValue = currentValue

    DDropdown {
        id: dropdown
        enabled: root.enabled
        Accessible.name: root.text
        Accessible.description: root.description + (root.description ? " · " : "") + currentValue
        width: Math.min(dropdownWidth, root.width - SettingsMetrics.rowPaddingH * 2)
        triggerHeight: Theme.buttonHeightXS
        triggerRadius: Theme.cornerRadiusS
        backgroundColor: Theme.surfaceContainerLowest
        hoverBackgroundColor: Theme.surfaceContainerLow
        normalBorderColor: Theme.outlineVariant
        focusedBorderColor: Theme.primary
        Component.onCompleted: currentValue = root.currentValue
        onValueChanged: value => root.valueChanged(value)
    }
}
