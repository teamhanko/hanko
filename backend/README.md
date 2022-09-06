# Hanko backend

Hanko backend provides an HTTP API to build a modern login and registration experience for your users. Its core features are an API for passkeys (WebAuthn), passwords, and passcodes, as well as JWT management.

Hanko backend can be used on its own or in combination with [elements](../elements), a powerful frontend library that contains polished and customizable UI flows for password-based and passwordless user authentication that can be easily integrated into any web app with as little as two lines of code.

## API features

- Passkeys (WebAuthn)
- Passcodes
- Passwords
- Email verification
- JWT management
- User management

## Upcoming features

- Exponential backoff for password attempts and passcode email sending
- 2FA configurations (optional, mandatory)

## Basic usage

The easiest way to start Hanko backend service is through docker. But before we can do that, we need to create a config file.

> **Note** If you just want to jump right into the experience of passkeys and passcodes, head over to the [quickstart guide](../README.md#quickstart).

### Config

Create a file with the name `config.yaml` and paste the config from below. Fill out the parameters marked with `<CHANGE-ME>` and, if you have access to an SMTP server, fill out the corresponding parameters with the information of your SMTP server.

If you don't know how to fill the webauthn config, see [here](./docs/Config.md#web-authentication).

```yaml
database:
  user: <CHANGE-ME>
  password: <CHANGE-ME>
  host: <CHANGE-ME>
  port: <CHANGE-ME>
  dialect: postgres
passcode:
  email:
    from_address: no-reply@next-unicorn.io
  smtp:
    host: <CHANGE-ME>
    user: <CHANGE-ME>
    password: <CHANGE-ME>
secrets:
  keys:
    - <CHANGE-ME>
service:
  name: Next Unicorn Authentication Service
webauthn:
  relying_party:
    id: <CHANGE-ME>
    display_name: <CHANGE-ME>
    origin: <CHANGE-ME>
```

> **Note** You need to change the smtp config to start the service. You can enter any host, user and password,
> they will not be checked for correctness at startup. But be aware that no emails will be sent
> and your users might not be able to login if no valid smtp server is set up.

> **Note** `secrets.keys` must be a random generated string at least 16 characters long.

### Docker

#### Database migrations

Before you can start and use the service you need to run the database migrations:

```shell
docker run --mount type=bind,source=<PATH-TO-CONFIG-FILE>,target=/config/config.yaml -p 8000:8000 -it ghcr.io/teamhanko/hanko:main migrate up
```

> **Note** The `<PATH-TO-CONFIG-FILE>` must be an absolute path to your config file created above.

#### Start the service

To start the service just run:

```shell
docker run --mount type=bind,source=<PATH-TO-CONFIG-FILE>,target=/config/config.yaml -p 8000:8000 -it ghcr.io/teamhanko/hanko:main serve public
```

> **Note** The `<PATH-TO-CONFIG-FILE>` must be an absolute path to your config file created above.

The service is now available at `localhost:8000`.

### From source

#### Building

To build the Hanko backend you only need to have [go installed](https://go.dev/doc/install) on your computer.

```shell
go build -a -o hanko main.go
```

This command will create an executable with the name `hanko`, which then can be used to start the Hanko backend.

#### Database migrations

Before you can start and use the service you need to run the database migrations:

```shell
./hanko migrate up --config <PATH-TO-CONFIG-FILE>
```

> **Note** The path to the config file can be relative or absolute.

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

> **Warning** The private API must be protected by an access management system.

### Supported Databases

Hanko backend supports the following databases:
- CockroachDB
- MariaDB
- MySQL
- PostgreSQL

### Rate Limiting

Currently, Hanko backend does not implement rate limiting in any way. In production systems, you may want to hide the Hanko service
behind a proxy or gateway (e.g. Kong, Traefik) that provides rate limiting.

### Configuration

All available configuration parameters can be found [here](./docs/Config.md).

## API specification

The API specification can be found [here](https://teamhanko.github.io/hanko/).

## License
The hanko backend ist licensed under the [AGPL-3.0](LICENSE).
