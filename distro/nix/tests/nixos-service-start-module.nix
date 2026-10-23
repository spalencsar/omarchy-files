{
  self,
  pkgs,
  ...
}:
pkgs.testers.runNixOSTest {
  name = "omarchy-files-nixos-service-start-module";

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
      systemd = {
        enable = true;
        target = "default.target";
      };
    };

    system.stateVersion = "25.11";
  };

  testScript = ''
    machine.wait_for_unit("multi-user.target")

    unit = "/run/current-system/sw/share/systemd/user/omarchy-files.service"
    machine.succeed(f"test -f {unit}")
    machine.succeed(f"grep -q 'run --session --hidden --keep-running' {unit}")
    machine.succeed("test -e /etc/systemd/user/default.target.wants/omarchy-files.service")
  '';
}
