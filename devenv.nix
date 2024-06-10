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
    pkgs.cacert
  ] ++ lib.optionals ( !config.container.isBuilding) [
    pkgs.docker
    pkgs.docker-compose
    pkgs.git
  ];

  enterShell = ''
    if [[ ! -f .env ]]; then
      cp .env.template .env
      echo "Created a new .env file from .env.example"
    fi
  '';

  processes.serveBackend = {
    exec = "cd backend && ${goPkgs.go_1_20}/bin/go";
  };

  containers = {
    "hanko" = {
      copyToRoot = ./backend;
      name = config.env.IMAGE_NAME;
      startupCommand = ''
        export SSL_CERT_DIR=${pkgs.cacert}/etc/ssl/certs/
        ${goPkgs.go_1_20}/bin/go generate ./...
        CGO_ENABLED=0 GOOS=linux GOARCH="$TARGETARCH" ${goPkgs.go_1_20}/bin/go build -a -o hanko main.go
        ./hanko
      '';
    };
  };
}
