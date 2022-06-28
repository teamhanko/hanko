# Hanko backend

The Hanko API offers a set of functions to create a modern login and registration experience to your users. It works
well with the [<hanko-auth> element](../hanko-js/README.md), a web component that can be integrated in your Web App and
provides a modern login and registration UI.

## Features

- passcodes
- web authentication (passkeys)
- passwords
- email verification

## Upcoming features

- backoff mechanisms
- more testing and code documentation

## Basic Usage

The easiest way to start the service is through docker. But before we can do that, we need to create a config file.

> If you only want to feel the experience of passkeys and passcodes head over [here](../README.md#Quickstart).

### Config

Create a file with the name `config.yaml` and paste the config from below. Fill out the params marked
with `<PLEASE-CHANGE-ME>`and if you have access to an SMTP server fill out the corresponding params with the information
of your SMTP server.

```yaml
database:
  user: <PLEASE-CHANGE-ME>
  password: <PLEASE-CHANGE-ME>
  host: <PLEASE-CHANGE-ME>
  port: <PLEASE-CHANGE-ME>
  dialect: postgres
passcode:
  email:
    from_address: no-reply@next-unicorn.io
  smtp:
    host: <PLEASE-CHANGE-ME>
    user: <PLEASE-CHANGE-ME>
    password: <PLEASE-CHANGE-ME>
secrets:
  keys:
    - <PLEASE-CHANGE-ME>
service:
  name: Next-Unicorn Authentication Service
```

> **Note:** You need to change the smtp config to start the service. You can enter any host, user and password,
> they will not be checked at startup for correctness. But be aware, if they are incorrect, that no emails will be sent
> and you and your users might not be able to login.

> **Note:** `secrets.keys` must be a random generated string at least 16 characters long.

### Docker

#### Database migrations

Before you can start and use the service you need to run the database migrations:

```shell
docker run --mount type=bind,source=<PATH-TO-CONFIG-FILE>,target=/config/config.yaml -p 8000:8000 -it ghcr.io/teamhanko/hanko:main migrate up
```

> **Note:** The `<PATH-TO-CONFIG-FILE>` must be an absolute path to your config file created above.

#### Start the service

To start the service just run:

```shell
docker run --mount type=bind,source=<PATH-TO-CONFIG-FILE>,target=/config/config.yaml -p 8000:8000 -it ghcr.io/teamhanko/hanko:main serve public
```

> **Note:** The `<PATH-TO-CONFIG-FILE>` must be an absolute path to your config file created above.

The service is now available at `localhost:8000`.

### From source

#### Building

To build the Hanko API you only need to have [go installed](https://go.dev/doc/install) on your computer.

```shell
go build -a -o hanko main.go
```

This command will create an executable with the name `hanko`, which then can be used to start the Hanko API.

#### Database migrations

Before you can start and use the service you need to run the database migrations:

```shell
./hanko migrate up --config <PATH-TO-CONFIG-FILE>
```

> **Note:** The path to the config file can be relative or absolute.

#### Start the service

To start the service just run:

```shell
./hanko serve public --config <PATH-TO-CONFIG-FILE>
```

The service is now available at `localhost:8000`.

## Advanced Usage

### Start private API

In the usage section above we only started the public API. Use the command below to start the private API. The default
port is `8001`, but can be [customized](./docs/Config.md) in the config.

```shell
serve private
```

Use this command to start the public and private API together:

````shell
serve all
````

> :warning: The private API must be protected by an access management.

### Configuration

All available configuration params can be found [here](./docs/Config.md).

### Rate Limiting

The Hanko service does not implement rate limiting in any way. So in production you want to hide the Hanko service
behind a proxy or gateway (e.g. kong, traeffik) which implements rate limiting.

## API specification

The API specification can be found [here](https://teamhanko.github.io/hanko/).
