import QtQuick
import Quickshell
import Quickshell.Io
import qs.Ui

BarWidget {
  id: root

  moduleName: "io.github.spalencsar.omarchy-files"

  property string binaryPath: ""

  readonly property string homePath: Quickshell.env("HOME")
  readonly property string localBinaryPath: homePath + "/.local/bin/omarchy-files"
  readonly property string installerPath: decodeURIComponent(
    String(Qt.resolvedUrl("install-omarchy.sh")).replace("file://", ""))
  readonly property bool installed: binaryPath !== ""

  function refresh() {
    if (!localCheck.running) localCheck.running = true
  }

  function install() {
    Quickshell.execDetached([
      "omarchy-launch-floating-terminal-with-presentation",
      installerPath
    ])
  }

  function launch() {
    if (!installed) {
      install()
      return
    }

    Quickshell.execDetached(["uwsm-app", "--", binaryPath, "show"])
  }

  implicitWidth: button.implicitWidth
  implicitHeight: button.implicitHeight

  Component.onCompleted: refresh()

  Timer {
    interval: 3000
    running: true
    repeat: true
    onTriggered: root.refresh()
  }

  Process {
    id: localCheck
    command: ["/usr/bin/test", "-x", root.localBinaryPath]
    onExited: function(exitCode) {
      if (exitCode === 0) {
        root.binaryPath = root.localBinaryPath
        return
      }
      if (!systemCheck.running) systemCheck.running = true
    }
  }

  Process {
    id: systemCheck
    command: ["/usr/bin/test", "-x", "/usr/bin/omarchy-files"]
    onExited: function(exitCode) {
      if (exitCode === 0) {
        root.binaryPath = "/usr/bin/omarchy-files"
        return
      }
      if (!localSystemCheck.running) localSystemCheck.running = true
    }
  }

  Process {
    id: localSystemCheck
    command: ["/usr/bin/test", "-x", "/usr/local/bin/omarchy-files"]
    onExited: function(exitCode) {
      root.binaryPath = exitCode === 0 ? "/usr/local/bin/omarchy-files" : ""
    }
  }

  BarIconButton {
    id: button
    anchors.fill: parent
    bar: root.bar
    tooltipText: root.installed
      ? "Open Omarchy Files"
      : "Install Omarchy Files"
    iconComponent: Component {
      Image {
        source: Qt.resolvedUrl("assets/io.github.spalencsar.OmarchyFiles.svg")
        fillMode: Image.PreserveAspectFit
        sourceSize.width: Math.round(width * Screen.devicePixelRatio)
        sourceSize.height: Math.round(height * Screen.devicePixelRatio)
      }
    }
    onPressed: function(buttonCode) {
      if (buttonCode === Qt.LeftButton) root.launch()
    }
  }
}
