{
  description = "bgenix";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = {nixpkgs, ...}: let
    systems = ["x86_64-linux" "aarch64-linux"];
    forEachSystem = nixpkgs.lib.genAttrs systems;
    pkgsForEach = nixpkgs.legacyPackages;
  in {
    packages = forEachSystem (system: {
      default = pkgsForEach.${system}.callPackage ./default.nix {};
    });

    devShells = forEachSystem (system: {
      default = pkgsForEach.${system}.callPackage ./shell.nix {};
    });
  };
}
