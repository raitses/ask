{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    git
  ];

  shellHook = ''
    echo "Go development environment loaded"
    echo "Go version: $(go version)"
  '';
}
