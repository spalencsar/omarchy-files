{
  lib,
  dfilesPkgs,
  pkgs,
  ...
}:
let
  inherit (lib) types;
in
{
  options.programs.omarchy-files = {
    enable = lib.mkEnableOption "Omarchy Files";

    package = lib.mkPackageOption dfilesPkgs "omarchy-files" {
      extraDescription = "The Omarchy Files package to use (defaults to be built from source)";
    };

    quickshell = {
      package = lib.mkPackageOption pkgs "quickshell" {
        extraDescription = "The Quickshell package used to launch the file manager UI";
      };
    };

    portal = {
      enable = lib.mkEnableOption "serving open and save dialogs through xdg-desktop-portal (registers the omarchy-files backend)";

      preferred = lib.mkOption {
        type = types.bool;
        default = false;
        description = "Prefer omarchy-files for org.freedesktop.impl.portal.FileChooser in the portal configuration named by portal.desktop.";
      };

      desktop = lib.mkOption {
        type = types.str;
        default = "common";
        description = "Which xdg.portal.config entry receives the FileChooser preference (a lowercased XDG_CURRENT_DESKTOP name, or common).";
      };
    };

    systemd = {
      enable = lib.mkEnableOption "Omarchy Files systemd startup";

      target = lib.mkOption {
        type = types.str;
        default = "graphical-session.target";
        description = "Systemd target that will automatically start omarchy-files.";
      };

      restartIfChanged = lib.mkOption {
        type = types.bool;
        default = true;
        description = "Auto-restart omarchy-files.service when omarchy-files changes";
      };
    };
  };
}
