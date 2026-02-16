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
  - [Social connections](#social-connections)
    - [Built-in providers](#built-in-providers)
    - [Custom OAuth/OIDC providers](#custom-oauthoidc-providers)
    - [Account linking](#account-linking)
  - [User metadata](#user-metadata)
  - [User import](#user-import)
  - [Webhooks](#webhooks)
  - [Session JWT templates](#session-jwt-templates)
- [API specification](#api-specification)
- [Configuration reference](#configuration-reference)
- [License](#license)

## API features

- Passkeys (WebAuthn)
- Passcodes
- Passwords
- Email verification
- 2FA (TOTP, security keys)
- JWT management
- Sessions
- User management
- OAuth/OIDC SSO identity providers
- SAML
- Webhooks

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
email_delivery:
  enabled: true
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
you need to supply the `email_delivery.smtp.host`, `email_delivery.smtp.port` as well as the `email_delivery.smtp.user`,
`email_delivery.smtp.password` settings according to your server/service settings.

### Configure JSON Web Key Set generation

The API uses [JSON Web Tokens](https://www.rfc-editor.org/rfc/rfc7519.html) (JWTs) for
[authentication](https://docs.hanko.io/api-reference/public/introduction).
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
[endpoint](https://docs.hanko.io/api-reference/public/well-known/get-json-web-key-set) to enable clients to verify token
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

`8000` is the default port for the public API. It can
be [customized](https://github.com/teamhanko/hanko/wiki/hanko-properties-server-properties-public#address) in the
configuration through the `server.public.address` option.

The service is now available at `localhost:8000`.

#### Start the admin API

In the usage section above we only started the public API. Use the command below to start the admin API. The default
port is `8001`, but can be
[customized](https://github.com/teamhanko/hanko/wiki/hanko-properties-server-properties-admin) in the configuration
through the `server.admin.address` option.

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

### Social connections

Hanko supports OAuth-based ([authorization code flow](https://www.rfc-editor.org/rfc/rfc6749#section-1.3.1)) third
party provider logins. The `third_party` configuration
[option](https://github.com/teamhanko/hanko/wiki/config-properties-third_party) contains all relevant configuration.
This includes options for setting up redirect URLs (in case of success or error on authentication with a provider) that
apply to both [built-in](#built-in-providers) and
[custom](#custom-oauthoidc-providers) providers.


#### Built-in providers

Built-in providers can be configured through the `third_party.providers` configuration [option](https://github.com/teamhanko/hanko/wiki/config-properties-third_party).
They must be explicitly `enabled` (i.e. providers are disabled default).
All provider configurations require provider credentials in the form of a client ID (`client_id`)
and a client secret (`secret`). See the guides in the official documentation for instructions on how to obtain these:

- [Apple](https://docs.hanko.io/guides/authentication-methods/oauth/apple)
- [Discord](https://docs.hanko.io/guides/authentication-methods/oauth/discord)
- [GitHub](https://docs.hanko.io/guides/authentication-methods/oauth/github)
- [Google](https://docs.hanko.io/guides/authentication-methods/oauth/google)
- [LinkedIn](https://docs.hanko.io/guides/authentication-methods/oauth/linkedin)
- [Microsoft](https://docs.hanko.io/guides/authentication-methods/oauth/microsoft)

#### Custom OAuth/OIDC providers

Custom providers can be configured through the `third_party.custom_providers` configuration
[option](https://github.com/teamhanko/hanko/wiki/config-properties-third_party-properties-custom_providers).
Like built-in providers they must be explicitly `enabled` and require a `client_id` and `secret`, which must
be obtained from the respective provider.
Custom providers can use either OAuth or OIDC. OIDC providers can be configured to use
[OIDC Discovery](https://openid.net/specs/openid-connect-discovery-1_0.html) by setting the `use_discovery`
option to `true`. An `issuer` must be configured too in that case. Otherwise both OAuth and OIDC providers
can manually define required endpoints (`authorization_endpoint`, `token_endpoint`, `userinfo_endpoint`).
`scopes` must be explicitly defined (with `openid` being the minimum requirement in case of OIDC providers).

#### Account linking

The `allow_linking` configuration option for built-in and custom providers determines whether automatic account linking for this provider
is activated. Note that account linking is based on e-mail addresses and OAuth providers may allow account holders to
use unverified e-mail addresses or may not provide any information at all about the verification status of e-mail
addresses. This poses a security risk and potentially allows bad actors to hijack existing Hanko
accounts associated with the same address. It is therefore recommended to make sure you trust the provider and to
also enable `emails.require_verification` in your configuration to ensure that only verified third party provider
addresses may be used.

### User metadata

Hanko allows for defining arbitrary user metadata. Metadata can be categorized into
three types that differ as to how they can be accessed and modified:

| Metadata type | Public API                   | Admin API             |
|---------------|------------------------------|-----------------------|
| Private       | No read or write access      | Read and write access |
| Public        | Read access                  | Read and write access |
| Unsafe        | Read access and write access | Read and write access |

Each metadata type supports a maximum of 3,000 characters. Metadata is stored as compact JSON (whitespace is ignored).
JSON syntax characters (`{`, `:`, `"`, `}`) count toward the character limit.
Multibyte UTF-8 characters (like emojis or non-Latin characters) count as 1 character each.

#### Private metadata

Private metadata should be used for sensitive data that should not be exposed to the client (e.g., internal flags/ids,
configuration, or access control details).

Private metadata can be read through the Admin API only using the
[Get metadata of a user](/api-reference/admin/user-management/get-metadata-of-a-user)
endpoint.

Private metadata can be set and modified through the Admin API only by using the
[Patch metadata of a user](https://docs.hanko.io/api-reference/admin/user-management/patch-metadata-of-a-user) endpoint.

#### Public metadata

Public metadata should be used for non-sensitive information that you want accessible but not modifiable by the client
(e.g., certain user roles, UI preferences, display options).

Public metadata can be read through the Public API, the Admin API and in JWT templates for customizing
the session JWT:

- `Public API`:
  - Public metadata is returned in the `user` object in the payload on the `success` state in a
    [Login](https://docs.hanko.io/api-reference/flow/login) and
    [Registration](https://docs.hanko.io/api-reference/flow/registration) flow as well
    as in the payload on the `profile_init` state in a [Profile](https://docs.hanko.io/api-reference/flow/profile) flow.
  - Public metadata is returned as part of the response of the
    [Get a user by ID](https://docs.hanko.io/api-reference/public/user-management/get-a-user-by-id) endpoint.
- `Admin API`:
  - Public metadata is returned as part of the response of the
    [Get metadata of a user](https://docs.hanko.io/api-reference/admin/user-metadata-management/get-metadata-of-a-user)
    endpoint.
  - Public metadata is returned as part of the response of the
    [Get a user by ID](https://docs.hanko.io/api-reference/admin/user-management/get-a-user-by-id) endpoint.
- `JWT Templates`:
  - Public metadata can be accessed through the `User` context object available on session JWT customization.
    See [Session JWT templates](#session-jwt-templates) for more details.

Public metadata can be set and modified through the Admin API only by using the
[Patch metadata of a user](https://docs.hanko.io/api-reference/admin/user-management/patch-metadata-of-a-user) endpoint.

#### Unsafe metadata

Unsafe metadata should be used for non-sensitive, temporary or experimental data that doesn't need strong safety
guarantees.

Unsafe metadata can be read through the Public API, the Admin API and in JWT templates for customizing
the session JWT:

- `Public API`:
    - Unsafe metadata is returned in the `user` object in the payload on the `success` state in a
      [Login](https://docs.hanko.io/api-reference/flow/login) and
      [Registration](https://docs.hanko.io/api-reference/flow/registration) flow as well
      as in the payload on the `profile_init` state in a [Profile](https://docs.hanko.io/api-reference/flow/profile) flow.
    - Unsafe metadata is returned as part of the response of the
      [Get a user by ID](https://docs.hanko.io/api-reference/public/user-management/get-a-user-by-id) endpoint.
- `Admin API`:
    - Unsafe metadata is returned as part of the response of the
      [Get metadata of a user](https://docs.hanko.io/api-reference/admin/user-metadata-management/get-metadata-of-a-user)
      endpoint.
    - Unsafe metadata is returned as part of the response of the
      [Get a user by ID](https://docs.hanko.io/api-reference/admin/user-management/get-a-user-by-id) endpoint.
- `JWT Templates`:
    - Unsafe metadata can be accessed through the `User` context object available on session JWT customization.
      See [Session JWT templates](#session-jwt-templates) for more details.

Unsafe metadata can be set and modified through the Public API and the Admin API:

- `Public API`:
  - Unsafe metadata can be set using the `patch_metadata` action in the
    [Profile](https://docs.hanko.io/api-reference/flow/profile) flow.

- `Admin API`:
  - Unsafe metadata can be set using the
    [Patch metadata of a user](https://docs.hanko.io/api-reference/admin/user-management/patch-metadata-of-a-user)
    endpoint.



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


### Webhooks

Webhooks are an easy way to get informed about changes in your Hanko instance (e.g. user or email updates).
To use webhooks you have to provide an endpoint on your application which can process the events. Please be aware that your
endpoint need to respond with an HTTP status code 200. Else-wise the delivery of the event will not be counted as successful.

#### Events
When a webhook is triggered it will send you a **JSON** body which contains the event and a jwt.
The JWT contains 2 custom claims:

* **data**: contains the whole object for which the change was made. (e.g.: the whole user object when an email or user is changed/created/deleted)
* **evt**: the event for which the webhook was triggered

A typical webhook event looks like:

```json
{
  "token": "the-jwt-token-which-contains-the-data",
  "event": "name of the event"
}
```

To decode the webhook you can use the JWKs created in [Configure JSON Web Key Set generation](#configure-json-web-key-set-generation)

#### Event Types

Hanko sends webhooks for the following event types:

| Event                       | Triggers on                                                                                        |
|-----------------------------|----------------------------------------------------------------------------------------------------|
| user                        | user creation, user deletion, user update, email creation, email deletion, change of primary email |
| user.create                 | user creation                                                                                      |
| user.delete                 | user deletion                                                                                      |
| user.login                  | user login                                                                                         |
| user.update                 | user update, email creation, email deletion, change of primary email                               |
| user.update.email           | email creation, email deletion, change of primary email                                            |
| user.update.email.create    | email creation                                                                                     |
| user.update.email.delete    | email deletion                                                                                     |
| user.update.email.primary   | change of primary email                                                                            |
| user.update.username.create | username creation                                                                                  |
| user.update.username.delete | username deletion                                                                                  |
| user.update.username.update | change of username                                                                                 |
| email.send                  | an email was sent or should be sent                                                                |

As you can see, events can have subevents. You are able to filter which events you want to receive by either selecting
a parent event when you want to receive all subevents or selecting specific subevents.

#### Enabling Webhooks

You can activate webhooks by adding the following snippet to your configuration file:

```yaml
webhooks:
  enabled: true
  hooks:
    - callback: <YOUR WEBHOOK ENDPOINT>
      events:
        - user
```

### Session JWT templates

You can define custom claims that will be added to session JWTs through the `session.jwt_template.claims`
configuration option.

These claims are processed at JWT generation time and can include static values,
templated strings using Go's text/template syntax, or nested structures (maps and slices).

The template has access to user data via the `.User` field, which includes:
- `.User.UserID`: The user's unique ID (string)
- `.User.Email`: Email details (optional, with `.Address`, `.IsPrimary`, `.IsVerified`)
- `.User.Username`: The user's username (string, optional)
- `.User.Metadata`: The user's public and unsafe metadata (optional)
    - `.User.Metadata.Public`: The user's public metadata (object)
    - `.User.Metadata.Unsafe`: The user's unsafe metadata (object)

#### Accessing user metadata

`.User.Metadata.Public` and `.User.Metadata.Unsafe`  can be accessed and queried using
[GJSON Path Syntax](https://github.com/tidwall/gjson/blob/master/SYNTAX.md) (try it out in the
[playground](https://gjson.dev/)).

Assume that a user's public metadata consisted of the following data:

```json
{
    "display_name": "GamerDude",
    "favorite_games": [
        {
            "name": "Legends of Valor",
            "genre": "RPG",
            "playtime_hours": 142.3
        },
        {
            "name": "Space Raiders",
            "genre": "Sci-Fi Shooter",
            "playtime_hours": 87.6
        }
    ]
}
```

Then you could, for example, access this data in the following ways in your templates:

```yaml
display_name: '{{ .User.Metadata.Public "display_name" }}'
favorite_games: '{{ .User.Metadata.Public "favorite_games" }}'
favorite_games_with_playtime_over_100: '{{ .User.Metadata.Public "favorite_games.#(playtime_hours>100)" }}'
favorite_genres: '{{ .User.Metadata.Public "favorite_games.#.genre" }}'
```

> **Note**
>
> Ensure you use proper quoting when accessing metadata. `.User.Metadata.Public` and `.User.Metadata.Unsafe`
are function calls internally and the given path argument must be a string, so it must be double quoted.
If you use use double quotes for your entire claim template then the path argument must be escaped, i.e.:
`"{{ .User.Metadata.Public \"display_name\" }}"`


Example usage in YAML configuration:
```yaml
role: "user"                                           # Static value
user_email: "{{.User.Email.Address}}"                  # Templated string
is_verified: "{{.User.Email.IsVerified}}"              # Boolean from user data
metadata:                                              # Nested map
  greeting: "Hello {{.User.Username}}"
  source: '{{ .User.Metadata.Public "display_name" }}' # Data read from public metadata
  ui_theme: '{{ .User.Metadata.Unsafe "ui_theme" }}'   # Data read from unsafe metadata
scopes:                                                # Slice with templated value
    - "read"
    - "write"
    - "{{if .User.Email.IsVerified}}admin{{else}}basic{{end}}"
```

In this example:
- `role` is a static string ("user").
- `user_email` dynamically inserts the user's email address.
- `is_verified` inserts a boolean indicating email verification status.
- `metadata` is a nested map with a static `source` and a templated `greeting`.
- `scopes` is a slice combining static values and a conditional template.

Notes:
- Custom claims are added at the top level of the session token [payload](#jwt-payload).
- Claims with the following keys will be ignored because they are currently added to the JWT by default:
    - `sub`
    - `iat`
    - `exp`
    - `aud`
    - `iss`
    - `email`
    - `username`
    - `session_id`
- Templates must conform to valid [Go text/template syntax](https://pkg.go.dev/text/template). Invalid templates are
  logged and excluded from the generated token.
- Boolean strings ("true" or "false") from templates are automatically converted to actual booleans.

For more details on template syntax, see: https://pkg.go.dev/text/template

## API specification

- [Hanko Public API](https://docs.hanko.io/api-reference/public/introduction)
- [Hanko Admin API](https://docs.hanko.io/api-reference/admin/introduction)

## Configuration reference

- [Using configuration file](https://github.com/teamhanko/hanko/wiki/Using-configuration-file)
- [Using environment variables](https://github.com/teamhanko/hanko/wiki/Using-environment-variables)
- [Configuration reference](https://github.com/teamhanko/hanko/wiki/config)


## License

The Hanko backend ist licensed under the [AGPL-3.0](../LICENSE).
