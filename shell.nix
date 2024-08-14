{
  mkShell,
  go,
  gopls,
  delve,
  ...
}:
mkShell {
  name = "bgenix-shell";
  packages = [go gopls delve];
}
