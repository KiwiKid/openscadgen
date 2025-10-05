{
  description = "OpenSCAD and openscadgen - A tool for generating parametric 3D models";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        
        # Build openscadgen from source
        openscadgen = pkgs.buildGoModule rec {
          pname = "openscadgen";
          version = "0.1.0";
          
          src = ./.;
          
          # Go module dependencies
          vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
          
          # Build flags
          buildFlags = [ "-ldflags=-s -w" ];
          
          # Pre-build step to generate templ files
          preBuild = ''
            # Generate templ files
            ${pkgs.go}/bin/go run github.com/a-h/templ/cmd/templ generate
          '';
          
          # Meta information
          meta = with pkgs.lib; {
            description = "A tool for generating parametric 3D models from OpenSCAD files";
            homepage = "https://github.com/kiwikid/openscadgen";
            license = licenses.mit;
            maintainers = [ ];
            platforms = platforms.all;
          };
        };
        
        # Development shell with all tools
        devShell = pkgs.mkShell {
          buildInputs = with pkgs; [
            # Go development
            go
            gopls
            golangci-lint
            
            # Templ for template generation
            (pkgs.buildGoModule {
              pname = "templ";
              version = "0.3.924";
              src = pkgs.fetchFromGitHub {
                owner = "a-h";
                repo = "templ";
                rev = "v0.3.924";
                hash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
              };
              vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
              subPackages = [ "cmd/templ" ];
            })
            
            # Air for live reloading
            (pkgs.buildGoModule {
              pname = "air";
              version = "1.49.0";
              src = pkgs.fetchFromGitHub {
                owner = "cosmtrek";
                repo = "air";
                rev = "v1.49.0";
                hash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
              };
              vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
            })
            
            # Just for task running
            just
            
            # OpenSCAD
            openscad
            
            # Other useful tools
            git
            curl
            wget
          ];
          
          shellHook = ''
            echo "🔧 OpenSCAD + openscadgen development environment"
            echo "Available commands:"
            echo "  just build     - Build the application"
            echo "  just test      - Run tests"
            echo "  just generate  - Generate templ files"
            echo "  just dev       - Start development with live reload"
            echo "  openscad       - OpenSCAD 3D modeling tool"
            echo "  openscadgen    - Your parametric model generator"
          '';
        };
        
      in
      {
        # Default package
        packages.default = openscadgen;
        
        # Individual packages
        packages.openscadgen = openscadgen;
        packages.openscad = pkgs.openscad;
        
        # Development shell
        devShells.default = devShell;
        
        # Apps for easy running
        apps = {
          openscadgen = flake-utils.lib.mkApp {
            drv = openscadgen;
            exePath = "/bin/openscadgen";
          };
          openscad = flake-utils.lib.mkApp {
            drv = pkgs.openscad;
            exePath = "/bin/openscad";
          };
        };
        
        # Formatter for the flake
        formatter = pkgs.nixpkgs-fmt;
      });
}
