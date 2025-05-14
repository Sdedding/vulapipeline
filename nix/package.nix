{ lib, ... }:
{
  perSystem =
    { pkgs, self', ... }:
    let
      inherit (pkgs) python3;
    in
    {
      packages = {
        default = self'.packages.vula;

        vula = python3.pkgs.buildPythonApplication {
          pname = "vula";
          # SUGGESTION software does not know its version or place version in plaintext file instead
          version = lib.pipe ../vula/__version__.py [
            builtins.readFile
            (lib.removePrefix ''__version__ = "'')
            (lib.removeSuffix ''"'')
            lib.trim
          ];

          src = lib.fileset.toSource {
            root = ../.;
            fileset =
              with lib.fileset;
              unions [
                (fileFilter (file: file.hasExt "py") ../.)
                ../README.md
                # SUGGESTION this is unused in the Nix package
                ../configs
                # SUGGESTION this is unused in the Nix package but perhaps should be
                ../misc/images
                # SUGGESTION this is unused in the Nix package
                ../misc/linux-desktop
                ../misc/python3-vula.postinst
                ../pyproject.toml
                ../pytest.ini
                ../setup.cfg
              ];
          };

          # without removing `pyproject.toml` we don't end up with an executable.
          postPatch = ''
            rm pyproject.toml
            substituteInPlace vula/frontend/constants.py \
              --replace "IMAGE_BASE_PATH = '/usr/share/icons/vula/'" "IMAGE_BASE_PATH = '$out/share/icons/vula/'"
          '';

          propagatedBuildInputs = with python3.pkgs; [
            click
            cryptography
            dbus-python
            highctidh
            packaging
            pillow
            pydbus
            pygments
            pygobject3
            pynacl
            pyroute2
            pystray
            pyyaml
            qrcode
            rendez
            schema
            setuptools
            tkinter
            zeroconf
          ];

          buildInputs = [ pkgs.libayatana-appindicator ];
          nativeBuildInputs = [
            pkgs.wrapGAppsHook
            pkgs.gobject-introspection
          ];
          nativeCheckInputs = with python3.pkgs; [ pytestCheckHook ];

          postInstall = ''
            mkdir -p $out/share
            ln -s $out/${python3.sitePackages}/usr/share/icons $out/share
          '';

          meta = {
            description = "Automatic local network encryption";
            homepage = "https://vula.link/";
            license = lib.licenses.gpl3Only;
            maintainers = with lib.maintainers; [
              lorenzleutgeb
              mightyiam
              stepbrobd
            ];
            mainProgram = "vula";
          };
        };
      };
    };
}
