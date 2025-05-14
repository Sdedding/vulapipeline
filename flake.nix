{
  inputs = {
    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };
    git-hooks = {
      url = "github:cachix/git-hooks.nix";
      inputs = {
        flake-compat.follows = "dedupe_flake-compat";
        nixpkgs.follows = "nixpkgs";
        gitignore.follows = "dedupe_gitignore";
      };
    };
    import-tree.url = "github:vic/import-tree";
    make-shell = {
      url = "github:nicknovitski/make-shell";
      inputs.flake-compat.follows = "dedupe_flake-compat";
    };
    nixpkgs.url = "github:mightyiam/nixpkgs/vula-deps";
    treefmt = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  # For deduplication
  inputs = {
    dedupe_flake-compat.url = "github:edolstra/flake-compat";
    dedupe_gitignore = {
      url = "github:hercules-ci/gitignore.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = inputs: inputs.flake-parts.lib.mkFlake { inherit inputs; } (inputs.import-tree ./nix);
}
