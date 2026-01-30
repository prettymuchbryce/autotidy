{
  description = "Automatically watch and organize files using configurable rules";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    # Optional inputs for testing modules
    darwin = {
      url = "github:lnl7/nix-darwin";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    home-manager = {
      url = "github:nix-community/home-manager";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, flake-utils, gomod2nix, darwin, home-manager, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        buildGoApplication = gomod2nix.legacyPackages.${system}.buildGoApplication;
      in {
        packages.default = buildGoApplication {
          pname = "autotidy";
          version = "0.1.0";
          src = ./.;
          pwd = ./.;
          modules = ./gomod2nix.toml;

          meta = with pkgs.lib; {
            description = "Automatically watch and organize files using configurable rules";
            homepage = "https://github.com/prettymuchbryce/autotidy";
            license = licenses.mit;
            maintainers = [];
          };
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            gotools
            gomod2nix.packages.${system}.default
          ];
        };
      }
    ) // {
      # Shared option definitions for autotidy modules
      _autotidyOptions = { lib, pkgs }: {
        enable = lib.mkEnableOption "autotidy";

        package = lib.mkOption {
          type = lib.types.package;
          default = self.packages.${pkgs.system}.default;
          description = "The autotidy package to use";
        };
      };

      # Shared launchd configuration
      _launchdConfig = pkg: {
        Label = "com.autotidy.daemon";
        ProgramArguments = [ "${pkg}/bin/autotidy" "daemon" ];
        RunAtLoad = true;
        KeepAlive = true;
        ProcessType = "Background";
        StandardOutPath = "/tmp/autotidy.out.log";
        StandardErrorPath = "/tmp/autotidy.err.log";
      };

      # Shared systemd service configuration
      _systemdServiceConfig = pkg: {
        Type = "notify";
        ExecStart = "${pkg}/bin/autotidy daemon";
        Restart = "on-failure";
        RestartSec = 5;
        TimeoutStopSec = 10;
        # Security hardening
        PrivateTmp = true;
        NoNewPrivileges = true;
      };

      # NixOS module for the service
      nixosModules.default = { config, lib, pkgs, ... }:
        let
          cfg = config.services.autotidy;
        in {
          options.services.autotidy = self._autotidyOptions { inherit lib pkgs; };

          config = lib.mkIf cfg.enable {
            environment.systemPackages = [ cfg.package ];

            systemd.user.services.autotidy = {
              description = "autotidy daemon";
              wantedBy = [ "default.target" ];
              after = [ "graphical-session.target" ];
              serviceConfig = self._systemdServiceConfig cfg.package;
            };
          };
        };

      # nix-darwin module (system-level macOS service)
      darwinModules.default = { config, lib, pkgs, ... }:
        let
          cfg = config.services.autotidy;
        in {
          options.services.autotidy = self._autotidyOptions { inherit lib pkgs; };

          config = lib.mkIf cfg.enable {
            environment.systemPackages = [ cfg.package ];

            launchd.user.agents.autotidy = {
              serviceConfig = self._launchdConfig cfg.package;
            };
          };
        };

      # Home-manager module (supports both Linux and macOS)
      homeModules.default = { config, lib, pkgs, ... }:
        let
          cfg = config.services.autotidy;
        in {
          options.services.autotidy = self._autotidyOptions { inherit lib pkgs; };

          config = lib.mkIf cfg.enable (lib.mkMerge [
            { home.packages = [ cfg.package ]; }

            # Linux: use systemd user service
            (lib.mkIf pkgs.stdenv.isLinux {
              systemd.user.services.autotidy = {
                Unit = {
                  Description = "autotidy daemon";
                  After = [ "graphical-session.target" ];
                };
                Service = self._systemdServiceConfig cfg.package;
                Install.WantedBy = [ "default.target" ];
              };
            })

            # macOS: use launchd agent
            (lib.mkIf pkgs.stdenv.isDarwin {
              launchd.agents.autotidy = {
                enable = true;
                config = self._launchdConfig cfg.package;
              };
            })
          ]);
        };

      # Test configurations for CI
      nixosConfigurations.test = nixpkgs.lib.nixosSystem {
        system = "x86_64-linux";
        modules = [
          self.nixosModules.default
          ({ pkgs, ... }: {
            services.autotidy.enable = true;

            # Minimal NixOS configuration for testing
            boot.loader.grub.enable = false;
            fileSystems."/" = { device = "/dev/null"; fsType = "tmpfs"; };
            system.stateVersion = "24.05";
          })
        ];
      };

      darwinConfigurations.test = darwin.lib.darwinSystem {
        system = "aarch64-darwin";
        modules = [
          self.darwinModules.default
          ({ pkgs, ... }: {
            services.autotidy.enable = true;

            # Required darwin configuration
            system.stateVersion = 4;
            system.primaryUser = "test";
          })
        ];
      };

      homeConfigurations.test-linux = home-manager.lib.homeManagerConfiguration {
        pkgs = nixpkgs.legacyPackages.x86_64-linux;
        modules = [
          self.homeModules.default
          {
            services.autotidy.enable = true;
            home.username = "test";
            home.homeDirectory = "/home/test";
            home.stateVersion = "24.05";
          }
        ];
      };

      homeConfigurations.test-darwin = home-manager.lib.homeManagerConfiguration {
        pkgs = nixpkgs.legacyPackages.aarch64-darwin;
        modules = [
          self.homeModules.default
          {
            services.autotidy.enable = true;
            home.username = "test";
            home.homeDirectory = "/Users/test";
            home.stateVersion = "24.05";
          }
        ];
      };
    };
}
