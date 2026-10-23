{
  self,
  pkgs,
  ...
}:
let
  homeManagerNixosModule =
    (fetchTarball {
      url = "https://github.com/nix-community/home-manager/archive/c03e4752899e55705dfa63979abd885c582a5c48.tar.gz";
      sha256 = "0c3jm12hgqns8wlgid5pfyg5qn39lsf7adv6wv5swjh54mhi3qj0";
    })
    + "/nixos";
in
pkgs.testers.runNixOSTest {
  name = "omarchy-files-home-manager-module";

  nodes.machine =
    { ... }:
    {
      imports = [ homeManagerNixosModule ];

      users.users.duxlabs = {
        isNormalUser = true;
        createHome = true;
        home = "/home/duxlabs";
        extraGroups = [ "wheel" ];
      };

      home-manager.useGlobalPkgs = true;
      home-manager.useUserPackages = true;

      home-manager.users.duxlabs =
        { pkgs, ... }:
        {
          imports = [ self.homeModules.omarchy-files ];

          home.username = "duxlabs";
          home.homeDirectory = "/home/duxlabs";
          home.stateVersion = "25.11";

          programs.omarchy-files = {
            enable = true;
            systemd = {
              enable = true;
              target = "default.target";
            };

            settings = {
              themeMode = "dark";
              radiusStrength = 50;
            };
          };
        };

      system.stateVersion = "25.11";
    };

  testScript = ''
    import json

    machine.wait_for_unit("multi-user.target")

    machine.succeed("su -- duxlabs -c 'command -v omarchy-files'")
    machine.succeed("su -- duxlabs -c 'test -f ~/.config/systemd/user/omarchy-files.service'")
    machine.succeed("su -- duxlabs -c 'test -L ~/.config/systemd/user/default.target.wants/omarchy-files.service'")

    machine.succeed("su -- duxlabs -c 'test -f ~/.config/omarchy-files/ui-settings.json'")
    settings = json.loads(machine.succeed("su -- duxlabs -c 'cat ~/.config/omarchy-files/ui-settings.json'"))
    t.assertEqual(settings["themeMode"], "dark")
    t.assertEqual(settings["radiusStrength"], 50)
  '';
}
