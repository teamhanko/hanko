# Quickstart

This directory contains an example application that showcases

- integration of web component(s) provided through the`@teamhanko/hanko-elements` package (see [elements](../elements)).
- server-side validation of JWTs issued by the Hanko [API](../backend) for securing a custom backend/API

The example is used on https://example.hanko.io/.

## Run the quickstart

The fastest way to try out Hanko is with [docker-compose](https://www.docker.com/products/docker-desktop/).

Clone this repository:
```
git clone https://github.com/teamhanko/hanko.git
```

Then, in the newly created `hanko` folder, run:
```
docker compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart" up --build
```
> **Note**: Docker (Desktop) needs to be running in order for the command to run.

After the services are up and running, the login page can be viewed at `localhost:8888`. To receive emails without your
own SMTP server, we added [mailslurper](https://github.com/mailslurper/mailslurper) which will be available at `localhost:8080`.

> **Note**: Hanko services are not published to a registry yet and will be built locally before the services are started.
