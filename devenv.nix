{ pkgs, lib, config, inputs, ... }:
let 
    goPkgs = pkgs.callPackage inputs.go_1_20_revision { };
in 
{
  dotenv.enable = true;
  dotenv.filename = ".env";

  name = "hanko";

  packages = [
    goPkgs.go_1_20
    pkgs.git
    pkgs.nixpacks
  ];

  enterShell = ''
    if [[ ! -f .env ]]; then
      cp .env.template .env
      echo "Created a new .env file from .env.example"
    fi
  '';

  scripts = {
    build_image.exec = ''
      nixpacks build ./backend \
        --name ${config.env.IMAGE_NAME}
    '';
    build_push_image.exec = "build_image && docker push ${config.env.IMAGE_NAME}";
  };
}
