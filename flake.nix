{
  description = "Auxio - S3-compatible object storage with image processing";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    gomod2nix = {
      url = "github:tweag/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, gomod2nix }:
    let
      supportedSystems = [ "x86_64-linux" "aarch64-linux" ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
      version = "0.1.0";
    in {
      packages = forAllSystems (system:
        let
          pkgs = import nixpkgs {
            inherit system;
            overlays = [ gomod2nix.overlays.default ];
          };
        in {
          default = pkgs.buildGoApplication {
            pname = "auxio";
            version = version;
            src = ./.;
            modules = ./gomod2nix.toml;
            # govips binds libvips via CGo, so CGo must be on and pkg-config +
            # libvips must be present at build time.
            CGO_ENABLED = "1";
            nativeBuildInputs = [ pkgs.pkg-config ];
            buildInputs = [ pkgs.vips ];
            # Without this the dashboard is not embedded and /dashboard serves
            # the development notice page instead of the UI.
            tags = [ "release" ];
            ldflags = [
              "-s"
              "-w"
              "-X main.version=${version}"
              "-X main.commit=${self.shortRev or "dirty"}"
            ];
            preBuild = ''
              if [ ! -f internal/dashboard/dist/index.html ]; then
                echo "ERROR: Frontend not built. Run 'just frontend' first."
                exit 1
              fi
            '';
          };
        });

      devShells = forAllSystems (system:
        let
          pkgs = import nixpkgs {
            inherit system;
            overlays = [ gomod2nix.overlays.default ];
          };
        in {
          default = pkgs.mkShell {
            buildInputs = with pkgs; [
              go
              gopls
              gomod2nix.packages.${system}.default
              pkg-config
              vips
              just
              awscli2
              minio-client
            ];
          };
        });

      nixosModules.default = import ./nix/module.nix self;
    };
}
