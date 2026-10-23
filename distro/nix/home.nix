{
  config,
  pkgs,
  lib,
  ...
}@args:
let
  cfg = config.programs.omarchy-files;
  jsonFormat = pkgs.formats.json { };
in
{
  imports = [ (import ./options.nix args) ];

  options.programs.omarchy-files.settings = lib.mkOption {
    type = jsonFormat.type;
    default = { };
    description = "Omarchy Files UI settings written to ~/.config/omarchy-files/ui-settings.json.";
    example = lib.literalExpression ''
      {
        themeMode = "dark";
        radiusStrength = 50;
      }
    '';
  };

  config = lib.mkIf cfg.enable {
    programs.omarchy-files.systemd.target = lib.mkDefault config.wayland.systemd.target;

    home.packages = [
      cfg.package
      cfg.quickshell.package
    ];

    dbus.packages = [ cfg.package ];

    xdg.portal.extraPortals = lib.mkIf cfg.portal.enable [ cfg.package ];
    xdg.portal.config.${cfg.portal.desktop}."org.freedesktop.impl.portal.FileChooser" = lib.mkIf (cfg.portal.enable && cfg.portal.preferred) [ "omarchy-files" ];

    systemd.user.services.omarchy-files = lib.mkIf cfg.systemd.enable {
      Unit = {
        Description = "Omarchy Files";
        PartOf = [ cfg.systemd.target ];
        After = [ cfg.systemd.target ];
      };

      Service = {
        ExecStart = lib.getExe cfg.package + " run --session --hidden --keep-running";
        Restart = "on-failure";
        RestartSec = "2";
        Slice = "app.slice";
      };

      Install.WantedBy = [ cfg.systemd.target ];
    };

    xdg.configFile."omarchy-files/ui-settings.json" = lib.mkIf (cfg.settings != { }) {
      source = jsonFormat.generate "ui-settings.json" cfg.settings;
    };
  };
}
