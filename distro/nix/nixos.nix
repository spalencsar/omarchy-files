{
  config,
  pkgs,
  lib,
  ...
}@args:
let
  cfg = config.programs.omarchy-files;
in
{
  imports = [ (import ./options.nix args) ];

  config = lib.mkIf cfg.enable {
    environment.systemPackages = [
      cfg.package
      cfg.quickshell.package
    ];

    services.dbus.packages = [ cfg.package ];

    xdg.portal.extraPortals = lib.mkIf cfg.portal.enable [ cfg.package ];
    xdg.portal.config.${cfg.portal.desktop}."org.freedesktop.impl.portal.FileChooser" = lib.mkIf (cfg.portal.enable && cfg.portal.preferred) [ "omarchy-files" ];

    systemd.packages = lib.mkIf cfg.systemd.enable [ cfg.package ];

    systemd.user.services.omarchy-files = lib.mkIf cfg.systemd.enable {
      wantedBy = [ cfg.systemd.target ];
      restartIfChanged = cfg.systemd.restartIfChanged;
      path = [ cfg.quickshell.package ];
    };
  };
}
