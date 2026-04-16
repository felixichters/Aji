{
  description = "Aji";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            # Go toolchain
            go
            gopls
            gotools

            # Node toolchain for the client
            nodejs_20
            pnpm

            # Task runner
            just
          ];

          shellHook = ''
            echo "aji devShell — go $(${pkgs.go}/bin/go version | awk '{print $3}'), node $(${pkgs.nodejs_20}/bin/node --version)"
          '';
        };
      });
}
