{
  flake.modules.nixos.example.services.vula = {
    enable = true;
    openFirewall = true;
  };
}
