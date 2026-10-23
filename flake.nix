{
  description = "Omarchy Files";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    dank-qml-common = {
      url = "github:duxlabs-dev/dux-qml-common";
      flake = false;
    };
  };

  outputs =
    { self, nixpkgs, dank-qml-common, ... }:
    let
      goModVersion =
        let
          content = builtins.readFile ./core/go.mod;
          lines = builtins.filter builtins.isString (builtins.split "\n" content);
          goLines = builtins.filter (l: builtins.match "go [0-9]+\\..*" l != null) lines;
          matched =
            if goLines != [ ] then builtins.match "go ([0-9]+)\\.([0-9]+).*" (builtins.head goLines) else null;
        in
        if matched != null then
          {
            major = builtins.elemAt matched 0;
            minor = builtins.elemAt matched 1;
          }
        else
          {
            major = "1";
            minor = "26";
          };
      goForPkgs = pkgs: pkgs.${"go_${goModVersion.major}_${goModVersion.minor}"};

      forEachSystem =
        fn:
        nixpkgs.lib.genAttrs [ "aarch64-linux" "x86_64-linux" ] (
          system: fn system nixpkgs.legacyPackages.${system}
        );

      mkModuleWithDfilesPkgs =
        modulePath:
        args@{ pkgs, ... }:
        {
          imports = [
            (import modulePath (args // { dfilesPkgs = buildDfilesPkgs pkgs; }))
          ];
        };

      mkDfiles =
        pkgs:
        pkgs.lib.makeOverridable (
          {
            withPdfPreview ? true,
          }:
          let
            version = "2.0.0";
            thumbnailTools = with pkgs; [
              bubblewrap
              ffmpegthumbnailer
              poppler-utils
            ];
            pdfQmlImportPath = "${pkgs.qt6.qtwebengine}/${pkgs.qt6.qtbase.qtQmlPrefix}";
          in
          (pkgs.buildGoModule.override { go = goForPkgs pkgs; }) {
            inherit version;
            pname = "omarchy-files";
            src = ./.;
            modRoot = "core";
            vendorHash = "sha256-FiRXQZtakQYiPn0ZqK+k+rG5ZcoaDm16T5t6Ankq/h8=";

            subPackages = [ "cmd/omarchy-files" ];

            tags = [ "withshell" ];

            # Mirror `make -C core sync-shell`. The flake src excludes submodule
            # content, so the DCommon symlink is replaced with the pinned input.
            postPatch = ''
              rm -rf core/internal/shellembed/dist
              cp -r quickshell core/internal/shellembed/dist
              rm -f core/internal/shellembed/dist/DCommon
              cp -r ${dank-qml-common}/DCommon core/internal/shellembed/dist/DCommon
              chmod -R u+w core/internal/shellembed/dist/DCommon
              rm -rf core/internal/shellembed/dist/scripts \
                core/internal/shellembed/dist/tests \
                core/internal/shellembed/dist/.claude
              rm -f core/internal/shellembed/dist/.qmlls.ini \
                core/internal/shellembed/dist/translations/extract_translations.py
            '';

            ldflags = [
              "-s"
              "-w"
              "-X main.Version=${version}"
              "-X main.Commit=${self.shortRev or self.dirtyShortRev or "dirty"}"
            ];

            nativeBuildInputs = with pkgs; [
              installShellFiles
              makeWrapper
            ];

            postInstall = ''
              install -Dm644 ${./assets/io.github.spalencsar.OmarchyFiles.desktop} \
                $out/share/applications/io.github.spalencsar.OmarchyFiles.desktop
              install -Dm644 ${./assets/io.github.spalencsar.OmarchyFiles.svg} \
                $out/share/icons/hicolor/scalable/apps/io.github.spalencsar.OmarchyFiles.svg
              install -Dm644 ${./assets/io.github.spalencsar.OmarchyFiles.metainfo.xml} \
                $out/share/metainfo/io.github.spalencsar.OmarchyFiles.metainfo.xml

              install -Dm644 ${./assets/io.github.spalencsar.OmarchyFiles.policy} \
                $out/share/polkit-1/actions/io.github.spalencsar.OmarchyFiles.policy
              substituteInPlace $out/share/polkit-1/actions/io.github.spalencsar.OmarchyFiles.policy \
                --replace-fail /usr/bin/omarchy-files $out/bin/.omarchy-files-wrapped

              install -Dm644 ${./assets/dbus/io.github.spalencsar.OmarchyFiles.FileManager1.service} \
                $out/share/dbus-1/services/io.github.spalencsar.OmarchyFiles.FileManager1.service
              substituteInPlace $out/share/dbus-1/services/io.github.spalencsar.OmarchyFiles.FileManager1.service \
                --replace-fail /usr/bin/omarchy-files $out/bin/omarchy-files

              install -Dm644 ${./assets/portal/omarchy-files.portal} \
                $out/share/xdg-desktop-portal/portals/omarchy-files.portal
              install -Dm644 ${./assets/dbus/org.freedesktop.impl.portal.desktop.omarchy-files.service} \
                $out/share/dbus-1/services/org.freedesktop.impl.portal.desktop.omarchy-files.service
              substituteInPlace $out/share/dbus-1/services/org.freedesktop.impl.portal.desktop.omarchy-files.service \
                --replace-fail /usr/bin/omarchy-files $out/bin/omarchy-files

              install -Dm644 ${./assets/systemd/omarchy-files.service} \
                $out/lib/systemd/user/omarchy-files.service
              substituteInPlace $out/lib/systemd/user/omarchy-files.service \
                --replace-fail /usr/bin/omarchy-files $out/bin/omarchy-files

              installShellCompletion --cmd omarchy-files \
                --bash <($out/bin/omarchy-files completion bash) \
                --fish <($out/bin/omarchy-files completion fish) \
                --zsh <($out/bin/omarchy-files completion zsh)

              wrapProgram $out/bin/omarchy-files \
                --prefix PATH : ${pkgs.lib.makeBinPath thumbnailTools} ${pkgs.lib.optionalString withPdfPreview "--prefix NIXPKGS_QT6_QML_IMPORT_PATH : ${pdfQmlImportPath}"}
            '';

            meta = {
              description = "File manager for Omarchy";
              homepage = "https://github.com/spalencsar/omarchy-files";
              license = pkgs.lib.licenses.mit;
              mainProgram = "omarchy-files";
              platforms = pkgs.lib.platforms.linux;
            };
          }
        ) { };

      buildDfilesPkgs = pkgs: {
        omarchy-files = mkDfiles pkgs;
      };
    in
    {
      packages = forEachSystem (
        system: pkgs: {
          omarchy-files = mkDfiles pkgs;
          default = self.packages.${system}.omarchy-files;
        }
      );

      lib = { inherit mkDfiles buildDfilesPkgs; };

      homeModules.omarchy-files = mkModuleWithDfilesPkgs ./distro/nix/home.nix;

      homeModules.default = self.homeModules.omarchy-files;

      nixosModules.omarchy-files = mkModuleWithDfilesPkgs ./distro/nix/nixos.nix;

      nixosModules.default = self.nixosModules.omarchy-files;

      devShells = forEachSystem (
        system: pkgs: {
          default = pkgs.mkShell {
            GOTOOLCHAIN = "auto";

            buildInputs = with pkgs; [
              (goForPkgs pkgs)
              gopls
              delve
              go-tools
              gnumake

              quickshell

              prek
              uv
              python3

              nixd
              nil
            ];

            shellHook = ''
              touch quickshell/.qmlls.ini 2>/dev/null
              if [ ! -f .git/hooks/pre-commit ]; then prek install; fi
            '';
          };
        }
      );

      nixosTests = nixpkgs.lib.genAttrs [ "aarch64-linux" "x86_64-linux" ] (
        system:
        import ./distro/nix/tests {
          inherit self;
          pkgs = nixpkgs.legacyPackages.${system};
        }
      );
    };
}
