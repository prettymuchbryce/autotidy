# Nix

autotidy provides a Nix flake with modules for NixOS, nix-darwin, and home-manager.

## home-manager

Works on both Linux and macOS:

```nix
{
  inputs.autotidy.url = "github:prettymuchbryce/autotidy";

  # Add to your home-manager modules:
  imports = [ inputs.autotidy.homeModules.default ];

  services.autotidy.enable = true;
}
```

## NixOS

```nix
{
  inputs.autotidy.url = "github:prettymuchbryce/autotidy";

  # Add to your NixOS modules:
  imports = [ inputs.autotidy.nixosModules.default ];

  services.autotidy.enable = true;
}
```

## nix-darwin

```nix
{
  inputs.autotidy.url = "github:prettymuchbryce/autotidy";

  # Add to your darwin modules:
  imports = [ inputs.autotidy.darwinModules.default ];

  services.autotidy.enable = true;
}
```

## Verify installation

```bash
autotidy status
```
