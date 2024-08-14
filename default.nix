{
  lib,
  # Dependencies
  age,
  jq,
  nix,
  mktemp,
  diffutils,
  buildGoModule,
  # Overridables
  ageBin ? "${age}/bin/age",
}: let
  pname = "bgenix";
  version = "0.1.0";
in
  buildGoModule {
    inherit pname version;
    src = lib.cleanSource ./.;

    # This is cringe.
    postPatch = ''
      substituteInPlace ./internal/config/constants.go \
        --replace-fail @nixInstantiate@ "${nix}/bin/nix-instantiate" \
        --replace-fail @ageBin@ "${ageBin}" \
        --replace-fail @jqBin@ "${jq}/bin/jq" \
        --replace-fail @mktempBin@ "${mktemp}/bin/mktemp" \
        --replace-fail @diffBin@ "${diffutils}/bin/diff" \
        --replace-fail @pname@ "${pname}" \
        --replace-fail @version@ "${version}"
    '';

    vendorHash = null;

    ldflags = ["-s" "-w"];

    meta.mainProgram = "bgenix";
  }
