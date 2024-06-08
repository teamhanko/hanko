{ pkgs, lib, config, inputs, ... }:
let 
    goPkgs = import (builtins.fetchGit {
         name = "go-1.20 Nixpkgs Version";
         url = "https://github.com/NixOS/nixpkgs/";
         ref = "refs/heads/nixpkgs-unstable";
         rev = "336eda0d07dc5e2be1f923990ad9fdb6bc8e28e3";
     }) {};
in 
{
  dotenv.enable = true;
  name = "hanko";

  languages = {
      go.enable = true;
      go.package = goPkgs.go_1_20;
  };

  processes.serveBackend = {
    exec = "cd backend && ${pkgs.go_1_20}/bin/go";
  };

  containers = {
    "hanko" = {
      copyToRoot = ./backend;
      name = config.IMAGE_NAME;
    };
  };
}
