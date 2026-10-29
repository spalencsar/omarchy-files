pragma ComponentBehavior: Bound

import QtQuick
import Quickshell
import Quickshell.Io
import qs.Ui as Ui
import "Version.js" as Version

Ui.BarWidget {
  id: root

  moduleName: "io.github.spalencsar.omarchy-files"

  property string binaryPath: ""
  property string installedVersion: ""

  readonly property string expectedVersion: "@APP_VERSION@"
  readonly property string homePath: Quickshell.env("HOME")
  readonly property string localBinaryPath: homePath + "/.local/bin/omarchy-files"
  readonly property string installerPath: decodeURIComponent(
    String(Qt.resolvedUrl("install-omarchy.sh")).replace("file://", ""))
  readonly property bool installed: binaryPath !== ""
  readonly property bool pluginManaged: binaryPath === localBinaryPath
  readonly property bool updateAvailable: pluginManaged
    && installedVersion !== ""
    && Version.compare(installedVersion, expectedVersion) < 0

  function useBinary(path) {
    if (binaryPath !== path) {
      binaryPath = path
      installedVersion = ""
    }
    if (path !== "" && installedVersion === "" && !versionCheck.running) {
      versionCheck.command = [path, "version"]
      versionCheck.running = true
    }
  }

  function checkVersion() {
    if (binaryPath === "" || versionCheck.running) return
    versionCheck.command = [binaryPath, "version"]
    versionCheck.running = true
  }

  function refresh() {
    if (!localCheck.running) localCheck.running = true
  }

  function install() {
    Quickshell.execDetached([
      "omarchy-launch-floating-terminal-with-presentation",
      installerPath
    ])
  }

  function activate() {
    if (!installed || updateAvailable) {
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

  Timer {
    interval: 30000
    running: root.installed
    repeat: true
    onTriggered: root.checkVersion()
  }

  Process {
    id: localCheck
    command: ["/usr/bin/test", "-x", root.localBinaryPath]
    onExited: function(exitCode) {
      if (exitCode === 0) {
        root.useBinary(root.localBinaryPath)
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
        root.useBinary("/usr/bin/omarchy-files")
        return
      }
      if (!localSystemCheck.running) localSystemCheck.running = true
    }
  }

  Process {
    id: localSystemCheck
    command: ["/usr/bin/test", "-x", "/usr/local/bin/omarchy-files"]
    onExited: function(exitCode) {
      root.useBinary(exitCode === 0 ? "/usr/local/bin/omarchy-files" : "")
    }
  }

  Process {
    id: versionCheck

    stdout: StdioCollector {
      waitForEnd: true
      onStreamFinished: root.installedVersion = Version.fromOutput(text)
    }

    onExited: function(exitCode) {
      if (exitCode !== 0) root.installedVersion = ""
    }
  }

  Ui.BarIconButton {
    id: button
    anchors.fill: root
    bar: root.bar
    tooltipText: root.updateAvailable
      ? "Update Omarchy Files to v" + root.expectedVersion
      : root.installed
        ? "Open Omarchy Files"
        : "Install Omarchy Files"
    iconComponent: Component {
      Item {
        Image {
          anchors.fill: parent
          source: Qt.resolvedUrl("assets/io.github.spalencsar.OmarchyFiles.svg")
          fillMode: Image.PreserveAspectFit
          sourceSize.width: Math.round(width * Screen.devicePixelRatio)
          sourceSize.height: Math.round(height * Screen.devicePixelRatio)
        }

        Rectangle {
          visible: root.updateAvailable
          anchors.right: parent.right
          anchors.bottom: parent.bottom
          width: Math.max(6, Math.round(parent.width * 0.28))
          height: width
          radius: width / 2
          color: button.foreground
        }
      }
    }
    onPressed: function(buttonCode) {
      if (buttonCode === Qt.LeftButton) root.activate()
    }
  }
}
