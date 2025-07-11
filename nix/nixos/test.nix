{ config, lib, ... }:
{
  perSystem =
    { pkgs, ... }:
    {
      checks.nixos = pkgs.nixosTest {
        name = "vula";

        nodes = {
          a = {
            imports = [ config.flake.modules.nixos.default ];

            services.vula = {
              enable = true;
              openFirewall = true;
              logLevel = "DEBUG";
              userPrefix = "non-default-prefix";
              systemGroup = "non-default-system-group";
              operatorsGroup = "vula-managers";
            };

            # Make sure that if hosts can resolve each others' names,
            # it is thanks to vula and the nss module it uses.
            networking.extraHosts = lib.mkForce "";

            users.users = {
              user.isNormalUser = true;
              admin = {
                isNormalUser = true;
                extraGroups = [ "vula-managers" ];
              };
            };
          };

          b.imports = [
            config.flake.modules.nixos.default
            config.flake.modules.nixos.example
          ];
        };

        testScript = ''
          start_all()
          a.wait_for_unit("vula-organize.service")
          a.wait_for_unit("vula-publish.service")
          a.wait_for_unit("vula-discover.service")
          b.wait_for_unit("vula-organize.service")
          b.wait_for_unit("vula-publish.service")
          b.wait_for_unit("vula-discover.service")

          def test_peer(node):
              peer = 'a.local.' if node.name == 'b' else 'b.local.'
              node.wait_until_succeeds(f"ping -I vula -c 1 {peer}", timeout=60)

              peer_ip = node.succeed(f"getent hosts {peer}").split()[0]
              route_result = node.succeed(f"ip route get {peer_ip}")
              assert " dev vula " in route_result

          test_peer(a)
          test_peer(b)

          a.succeed("pgrep --uid non-default-prefix-organize vula")
          a.succeed("pgrep --uid non-default-prefix-discover vula")
          a.succeed("pgrep --uid non-default-prefix-publish vula")

          group_count = a.succeed("pgrep --count --group non-default-system-group vula").strip()
          assert group_count == "3", "vula process group count should be 3"

          # log level
          a.succeed("pgrep --full -- 'vula.* --verbose organize'")
          a.succeed("pgrep --full -- 'vula.* --verbose discover'")
          a.succeed("pgrep --full -- 'vula.* --verbose publish'")

          a.fail("su - user -c 'vula status'")
          a.succeed("su - admin -c 'vula status'")

          a.fail("su - user -c 'vula peer'")
          a.succeed("su - admin -c 'vula peer'")

          a.fail("su - user -c 'vula prefs set pin_new_peers true'")
          a.succeed("su - admin -c 'vula prefs set pin_new_peers true'")
        '';

        interactive.nodes.b = {
          virtualisation = {
            memorySize = 4096;
            cores = 3;
          };
          users.users = {
            joe = {
              isNormalUser = true;
              password = "";
            };
            admin = {
              isNormalUser = true;
              extraGroups = [ "vula-ops" ];
              password = "";
            };
          };
          services.xserver = {
            enable = true;
            displayManager.gdm.enable = true;
            desktopManager.xfce.enable = true;
          };
        };
      };
    };
}
