{
  description = "dunk - run terminal coding agents in persistent cloud sandboxes";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      systems = [ "aarch64-darwin" "x86_64-darwin" "x86_64-linux" "aarch64-linux" ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f (import nixpkgs { inherit system; }));
    in {
      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          packages = with pkgs; [
            go
            gopls
            gotools
            golangci-lint
            git
            tmux
            nodejs_22
          ];
          shellHook = ''
            echo "dunk dev shell"
            echo "E2B CLI: brew install e2b  # or npm i -g @e2b/cli"
            echo "Run: go run ./cmd/dunk --help"
          '';
        };
      });

      packages = forAllSystems (pkgs: {
        default = pkgs.buildGoModule {
          pname = "dunk";
          version = "0.1.0";
          src = ./.;
          subPackages = [ "cmd/dunk" ];
          vendorHash = "sha256-komX1AmHt2NoF1x6xsNa2RFkfVzOXfYEMPhT0zwMxjw=";
        };
      });

      apps = forAllSystems (pkgs: {
        default = {
          type = "app";
          program = "${self.packages.${pkgs.system}.default}/bin/dunk";
        };
      });

      checks = forAllSystems (pkgs: {
        tests = self.packages.${pkgs.system}.default;
      });
    };
}
