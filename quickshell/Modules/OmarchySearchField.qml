import qs.Common
import qs.DCommon.Widgets

DTextField {
    // A deliberate, bordered field instead of DCommon's full-radius search pill.
    cornerRadius: Theme.cornerRadiusS
    backgroundColor: Theme.surfaceContainerLowest
    normalBorderColor: Theme.outlineVariant
    focusedBorderColor: Theme.primary
    borderWidth: Theme.outlineWidth
    focusedBorderWidth: Theme.outlineWidthFocused
}
