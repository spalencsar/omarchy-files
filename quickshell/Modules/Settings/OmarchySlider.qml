import QtQuick
import QtQuick.Controls as Controls
import qs.Common

Controls.Slider {
    id: root
    property alias minimum: root.from
    property alias maximum: root.to
    property alias step: root.stepSize
    property bool showStops: false
    property bool showValue: true
    property string size: "s"
    property string unit: ""
    property int decimals: 0
    signal sliderValueChanged(int newValue)

    implicitHeight: Theme.buttonHeightXS
    onMoved: sliderValueChanged(Math.round(value))
    onValueChanged: if (pressed) sliderValueChanged(Math.round(value))

    background: Rectangle {
        x: root.leftPadding
        y: root.topPadding + root.availableHeight / 2 - height / 2
        width: root.availableWidth
        height: 2
        color: Theme.outlineVariant
        Rectangle { width: parent.width * root.visualPosition; height: parent.height; color: Theme.primary }
    }
    handle: Rectangle {
        x: root.leftPadding + root.visualPosition * (root.availableWidth - width)
        y: root.topPadding + root.availableHeight / 2 - height / 2
        width: 8
        height: 14
        radius: Theme.cornerRadiusXS
        color: Theme.primary
        border.width: Theme.outlineWidth
        border.color: Theme.surface
    }
}
