{
  self,
  pkgs,
  ...
}:
pkgs.testers.runNixOSTest {
  name = "omarchy-files-nixos-module";

  nodes.machine = {
    imports = [
      self.nixosModules.omarchy-files
    ];

    users.users.duxlabs = {
      isNormalUser = true;
      extraGroups = [ "wheel" ];
    };

    programs.omarchy-files = {
      enable = true;
      systemd.enable = true;
    };

    system.stateVersion = "25.11";
  };

  testScript = ''
    machine.wait_for_unit("multi-user.target")

    machine.succeed("command -v omarchy-files")
    machine.succeed("command -v qs")
    machine.succeed("su -- duxlabs -c 'omarchy-files --help >/dev/null'")
    machine.succeed("test -f /run/current-system/sw/share/systemd/user/omarchy-files.service")
    machine.succeed("test -f /run/current-system/sw/share/dbus-1/services/io.github.spalencsar.OmarchyFiles.FileManager1.service")
    machine.succeed("test -f /run/current-system/sw/share/polkit-1/actions/io.github.spalencsar.OmarchyFiles.policy")
  '';
}
