{
  description = "jira-cli development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
        jira-cli = pkgs.buildGoModule {
          pname = "jira-cli";
          version = "1.0.0";
          src = ./.;
          vendorHash = "sha256-Orj3kQJYOM8kDjHGlNQVx4z3A/6661yQE+9J4SvPsWs=";
          subPackages = [ "cmd/jira" ];
        };
      in
      {
        packages = {
          default = jira-cli;
          jira-cli = jira-cli;
        };

        apps.default = flake-utils.lib.mkApp {
          drv = jira-cli;
          exePath = "/bin/jira";
        };

        devShells.default = pkgs.mkShell {
          packages = [
            pkgs.go
            pkgs.gotools
            pkgs.gopls
            pkgs.golangci-lint
            pkgs.goreleaser
            pkgs.gnumake
          ];

          shellHook = ''
            export CGO_ENABLED=1
          '';
        };
      });
}
