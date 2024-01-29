# Hanko backend

Hanko backend provides an HTTP API to build a modern login and registration experience for your users. Its core features
are an API for passkeys (WebAuthn), passwords, and passcodes, as well as JWT management.

Hanko backend can be used on its own or in combination with [hanko-elements](../frontend/elements), a powerful frontend library
that contains polished and customizable UI flows for password-based and passwordless user authentication that can be
easily integrated into any web app with as little as two lines of code.

# Contents

- [API features](#api-features)
- [Running the backend](#running-the-backend)
- [Running tests](#running-tests)
- [Additional topics](#additional-topics)
  - [Enabling password authentication](#enabling-password-authentication)
  - [Cross-domain communication](#cross-domain-communication)
  - [Audit logs](#audit-logs)
  - [Rate Limiting](#rate-limiting)
  - [Social logins](#social-logins)
  - [User import](#user-import)
- [API specification](#api-specification)
- [Configuration reference](#configuration-reference)
- [License](#license)

## API features

- Passkeys (WebAuthn)
- Passcodes
- Passwords
- Email verification
- JWT management
- User management

### Upcoming features

- Exponential backoff for password attempts and passcode email sending
- 2FA configurations (optional, mandatory)

## Running the backend

> **Note** If you just want to jump right into the experience of passkeys and passcodes, head over to the
> [quickstart guide](../quickstart/README.md).

To get the Hanko backend up and running you need to:

1. [Run a database](#run-a-database)
2. [Configure database access](#configure-database-access)
3. [Apply database migrations](#apply-database-migrations)
4. [Run and configure an SMTP server](#run-and-configure-an-smtp-server)
5. [Configure JSON Web Key Set generation](#configure-json-web-key-set-generation)
6. [Configure WebAuthn](#configure-webauthn)
7. [Configure CORS](#configure-cors)
8. [Start the backend](#start-the-backend)

### Run a database

The following databases are currently supported:

- PostgreSQL
- MySQL

#### Postgres

Use Docker to run a container based on the official [Postgres](https://hub.docker.com/_/postgres) image:

```shell
docker run --name=postgres \
-e POSTGRES_USER=<DB_USER> \
-e POSTGRES_PASSWORD=<DB_PASSWORD> \
-e POSTGRES_DB=<DB_DATABASE> \
-p <DB_PORT>:5432 \
-d postgres
```

or use the [official binary packages](https://www.postgresql.org/download/) to install and run
a Postgres instance.

#### MySQL

Use Docker to run a container based on the official [MySQL](https://hub.docker.com/_/mysql) image:

```shell
docker run --name=mysql \
-e MYSQL_USER=<DB_USER> \
-e MYSQL_PASSWORD=<DB_PASSWORD> \
-e MYSQL_DATABASE=<DB_DATABASE> \
-e MYSQL_RANDOM_ROOT_PASSWORD=true \
-p <DB_PORT>:3306 \
-d mysql:latest
```

or follow the official [installation instructions](https://dev.mysql.com/doc/mysql-getting-started/en/#mysql-getting-started-installing) to install and run
a MySQL instance.

### Configure database access

Open the `config.yaml` file in the `backend/config` or create your own `*.yaml` file and add the following:

```yaml
database:
  user: <DB_USER>
  password: <DB_PASSWORD>
  host: localhost # change this if the DB is not running on localhost, esp. in a production setting
  port: <DB_PORT>
  database: <DB_DATABASE>
  dialect: <DB_DIALECT> # depending on your choice of DB: postgres, mysql
```

Replace `<DB_USER>`, `<DB_PASSWORD>`, `<DB_PORT>`, `<DB_DATABASE>` with the values used in your running
DB instance (cf. the Docker commands above used for running the DB containers) and replace `<DB_DIALECT>` with
the DB of your choice.

### Apply Database migrations

Before you can start and use the service you need to run the database migrations:

#### Docker

```shell
docker run --mount type=bind,source=<PATH-TO-CONFIG-FILE>,target=/config/config.yaml -p 8000:8000 -it ghcr.io/teamhanko/hanko:latest migrate up
```

> **Note** The `<PATH-TO-CONFIG-FILE>` must be an absolute path to your config file created above.

#### From source

First build the Hanko backend. The only prerequisite is to have Go (v1.18+) [installed](https://go.dev/doc/install)
on your computer.

```shell
go generate ./...
go build -a -o hanko main.go
```

This command will create an executable with the name `hanko`, which then can be used to apply the database migrations
and start the Hanko backend.

To apply the migrations, run:

```shell
./hanko migrate up --config <PATH-TO-CONFIG-FILE>
```

> **Note** The path to the config file can be relative or absolute.


### Run and configure an SMTP server

The Hanko backend requires an SMTP server to send out mails containing
passcodes (e.g. for the purpose of email verification, password recovery).

For local development purposes you can use, e.g., [Mailslurper](https://www.mailslurper.com/).
Follow the official [installation](https://github.com/mailslurper/mailslurper/wiki/Getting-Started) instructions or
use an (inofficial) [Docker image](https://hub.docker.com/r/marcopas/docker-mailslurper) to get it up and running:

```shell
docker run --name=mailslurper -it -p 2500:2500 -p 8080:8080 -p 8085:8085 @marcopas/docker-mailslurper
```

where in this case
- `2500` is the SMTP port of the service
- `8080` is the port for the GUI application for managing mails
- `8085` is the port for the [API](https://github.com/mailslurper/mailslurper/wiki/API-Guide) service for managing mails

When using the above Docker command to run a Mailslurper container, it does not configure
a user/password, so a minimal configuration in your configuration file (`backend/config/config.yaml` or
your own `*.yaml` file) could contain the following:

```yaml
passcode:
  email:
    from_address: no-reply@example.com
    from_name: Example Application
  smtp:
    host: localhost
    port: 2500
```

To ensure that passcode emails also contain a proper subject header, configure a service
name:

```yaml
service:
  name: Example Authentication Service
```

In a production setting you would rather use a self-hosted SMTP server or a managed service like AWS SES. In that case
you need to supply the `passcode.smtp.host`, `passcode.smtp.port` as well as the `passcode.smtp.user`,
`passcode.smtp.password` settings according to your server/service settings.

> **Note** The `passcode.smtp.host` configuration entry is required for the service to start up.
> Only a check for a non-empty string value will be performed. Also: SMTP-connection related values are not
> verified, i.e. the application may start but no emails will be sent and your users might not be able to log in if
> the provided values do not describe an existing SMTP server.

### Configure JSON Web Key Set generation

The API uses [JSON Web Tokens](https://www.rfc-editor.org/rfc/rfc7519.html) (JWTs) for
[authentication](https://docs.hanko.io/api/public#section/Authentication).
JWTs are verified using [JSON Web Keys](https://www.rfc-editor.org/rfc/rfc7517) (JWK).
JWKs are created internally by setting `secrets.keys` options in the
configuration file (`backend/config/config.yaml` or your own `*.yaml` file):

```yaml
secrets:
  keys:
    - <CHANGE-ME>
```

> **Note**  at least one `secrets.keys` entry must be provided and each entry must be a random generated string at least 16 characters long.

Keys secrets are used to en- and decrypt the JWKs which get used to sign the JWTs.
For every key a JWK is generated, encrypted with the key and persisted in the database.

The Hanko backend API publishes public cryptographic keys as a JWK set through the `.well-known/jwks.json`
[endpoint](https://docs.hanko.io/api/public#tag/.well-known/operation/getJwks) to enable clients to verify token
signatures.

### Configure WebAuthn

Passkeys are based on the [Web Authentication API](https://www.w3.org/TR/webauthn-2/#web-authentication-api).
In order to create and login with passkeys, the Hanko backend must be provided information about
the [WebAuthn Relying Party](https://www.w3.org/TR/webauthn-2/#webauthn-relying-party).

For most use cases, you just need the domain of your web application that uses the Hanko backend. Set
`webauthn.relying_party.id` to the domain and set `webauthn.relying_party.origin` to the domain _including_ the
protocol.

> **Important**: If you are hosting your web application on a non-standard HTTP port (i.e. `80`) you also have to
> include this in the origin setting.

#### Local development example

When developing locally, the Hanko backend defaults to:

```yaml
webauthn:
  relying_party:
    id: "localhost"
    display_name: "Hanko Authentication Service"
    origins:
      - "http://localhost"
```

so no further configuration changes need to be made to your configuration file.

#### Production Examples

When you have a website hosted at `example.com` and you want to add a login to it that will be available
at `https://example.com/login`, the WebAuthn config would look like this:

```yaml
webauthn:
  relying_party:
    id: "example.com"
    display_name: "Example Project"
    origins:
      - "https://example.com"
```

If the login should be available at `https://login.example.com` instead, then the WebAuthn config would look like this:

```yaml
webauthn:
  relying_party:
    id: "login.example.com"
    display_name: "Example Project"
    origins:
      - "https://login.example.com"
```

Given the above scenario, you still may want to bind your users WebAuthn credentials to `example.com` if you plan to
add other services on other subdomains later that should be able to use existing credentials. Another reason can be if
you want to have the option to move your login from `https://login.example.com` to `https://example.com/login` at some
point. Then the WebAuthn config would look like this:

```yaml
webauthn:
  relying_party:
    id: "example.com"
    display_name: "Example Project"
    origins:
      - "https://login.example.com"
```

### Configure CORS

Because the backend and your application(s) consuming backend API most likely have different origins, i.e.
scheme (protocol), hostname (domain), and port part of the URL are different, you need to configure
Cross-Origin Resource Sharing (CORS) and specify your application(s) as allowed origins:

```yaml
server:
  public:
    cors:
      allow_origins:
        - https://example.com
```

When you include a wildcard `*` origin you need to set `unsafe_wildcard_origin_allowed: true`:

```yaml
server:
  public:
    cors:
      allow_origins:
        - "*"
      unsafe_wildcard_origin_allowed: true
```

Wildcard `*` origins can lead to cross-site attacks and when you include a `*` wildcard origin,
we want to make sure, that you understand what you are doing, hence this flag.

> **Note** In most cases, the `allow_origins` list here should contain the same entries as the `webauthn.relying_party.origins` list. Only when you have an Android app you will have an extra entry (`android:apk-key-hash:...`) in the `webauthn.relying_party.origins` list.

### Start the backend

The Hanko backend consists of a public and an administrative API (currently providing user management
endpoints). These can be started separately or in a single command.

#### Start the public API

##### Docker

```shell
docker run --mount type=bind,source=<PATH-TO-CONFIG-FILE>,target=/config/config.yaml -p 8000:8000 -it ghcr.io/teamhanko/hanko:latest serve public
```

##### Using pre-built binaries

Each [GitHub release](https://github.com/teamhanko/hanko/releases) (> 0.9.0) has `hanko`'s binary assets uploaded to it. Alternatively you can use
a tool like [eget](https://github.com/zyedidia/eget) to install binaries from releases on GitHub:

```bash
eget teamhanko/hanko
```

##### From source

```shell
go generate ./...
go build -a -o hanko main.go
```

Then run:
```shell
./hanko serve public --config <PATH-TO-CONFIG-FILE>
```

> **Note** The `<PATH-TO-CONFIG-FILE>` must be an absolute path to your config file created above.

`8000` is the default port for the public API. It can be [customized](./docs/Config.md) in the configuration through
the `server.public.address` option.

The service is now available at `localhost:8000`.

#### Start the admin API

In the usage section above we only started the public API. Use the command below to start the admin API. The default
port is `8001`, but can be [customized](./docs/Config.md) in the configuration through the
`server.admin.address` option.

```shell
serve admin
```

> **Warning** The admin API must be protected by an access management system.

##### Start both public and admin API

Use this command to start the public and admin API together:

```shell
serve all
```

## Running tests

You can run the unit tests by running the following command within the `backend` directory:

```bash
go test -v ./...
```

## Additional topics

### Enabling password authentication

Password-based authentication is disabled per default. You can activate it and set the minimum password
length in your configuration file:

```yaml
password:
  enabled: true
  min_password_length: 8
```

### Cross-domain communication

JWTs used for authentication are propagated via cookie. If your application and the Hanko backend run on different
domains, cookies cannot be set by the Hanko backend. In that case the backend must be configured to transmit the JWT via
Header (`X-Auth-Token`). To do so, enable propagation of the `X-Auth-Token` header:

```yaml
session:
  enable_auth_token_header: true
```

### Audit logs

API operations are recorded in an audit log. By default, the audit log is enabled
and logs to STDOUT:

```yaml
audit_log:
  console_output:
    enabled: true
    output: "stdout"
  storage:
    enabled: false
```

To persist audit logs in the database, set `audit_log.storage.enabled` to `true`.

### Rate Limiting

Hanko implements basic fixed-window rate limiting for the passcode/init and password/login endpoints to mitigate brute-force attacks.
It uses a combination of user-id/IP to mitigate DoS attacks on user accounts. You can choose between an in-memory and a redis store.

In production systems, you may want to hide the
Hanko service behind a proxy or gateway (e.g. Kong, Traefik) to provide additional network-based rate limiting.

### Social Logins

Hanko supports OAuth-based ([authorization code flow](https://www.rfc-editor.org/rfc/rfc6749#section-1.3.1)) third
party provider logins. See the `third_party` option in the [configuration reference](./docs/Config.md) on how to
configure them. All provider configurations require provider credentials. See the guides in the official
documentation for instructions on how to obtain these:

- [Apple](https://docs.hanko.io/guides/authentication-methods/oauth/apple)
- [GitHub](https://docs.hanko.io/guides/authentication-methods/oauth/github)
- [Google](https://docs.hanko.io/guides/authentication-methods/oauth/google)

#### Account linking

The `allow_linking` configuration option for providers determines whether automatic account linking for this provider
is activated. Note that account linking is based on e-mail addresses and OAuth providers may allow account holders to
use unverified e-mail addresses or may not provide any information at all about the verification status of e-mail
addresses. This poses a security risk and potentially allows bad actors to hijack existing Hanko
accounts associated with the same address. It is therefore recommended to make sure you trust the provider and to
also enable `emails.require_verification` in your configuration to ensure that only verified third party provider
addresses may be used.


### User import
You can import an existing user pool into Hanko using json in the following format:
```json
[
  {
    "user_id": "799e95f0-4cc7-4bd7-9f01-5fdc4fa26ea3",
    "emails": [
      {
        "address": "koreyrath@wolff.name",
        "is_primary": true,
        "is_verified": true
      }
    ],
    "created_at": "2023-06-07T13:42:49.369489Z",
    "updated_at": "2023-06-07T13:42:49.369489Z"
  },
  {
    "user_id": "",
    "emails": [
      {
        "address": "joshuagrimes@langworth.name",
        "is_primary": true,
        "is_verified": true
      }
    ],
    "created_at": "2023-06-07T13:42:49.369494Z",
    "updated_at": "2023-06-07T13:42:49.369494Z"
  }
]
```
There is a json schema file located [here](json_schema/hanko.user_import.json) that you can use for validation and input suggestions.
To import users run:

> hanko user import -i ./path/to/import_file.json


## API specification

- [Hanko Public API](https://docs.hanko.io/api/public)
- [Hanko Admin API](https://docs.hanko.io/api/admin)

## Configuration reference

[Configuration reference](./docs/Config.md)

## License

The Hanko backend ist licensed under the [AGPL-3.0](../LICENSE).
