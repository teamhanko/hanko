# Hanko backend

## Usage

### Serving api
#### Public
```shell
go run main.go serve public
```

#### Private
```shell
go run main.go serve private
```

#### Both
```shell
go run main.go serve all
```

### Applying migrations
#### Ups
```shell
go run main.go migrate up
```

#### Downs
```shell
go run main.go migrate down 1
```

## Config

On default the config file will be loaded from `./config/config.yaml`. But on startup you can specify a different location
to your configuration file with `--config` flag:

```shell
go run main.go serve all --config /etc/different/directory/config.yaml
```

### All config keys
```yaml
server:
  public:
    address: ":8000"
    cors:
      enabled: false
      allow_credentials: false
      allow_origins:
        - "*"
      allow_methods:
        - ""
      allow_headers:
        - ""
      expose_headers:
        - ""
      max_age: 0
  private:
    address: ":8001"
database:
  host: "localhost"
  port: "5432"
  dialect: "postgres"
  user: "hanko"
  password: "hanko"
  database: "hanko"
service:
  name: "Hanko Authentication Service"
secrets:
  keys:
    - "change-me"
session:
  lifespan: "1h"
  cookie:
    domain: ""
    http_only: true
    same_site: "strict"
password:
  enabled: false
passcode:
  ttl: 0
  email:
    from_address: ""
    from_name: ""
  smtp:
    host: ""
    port: ""
    user: ""
    password: ""
webauthn:
  timeout: 0
  relying_party:
    id: "localhost"
    display_name: ""
    origin: "http://localhost"
```

## API specification

The API specification can be found [here](https://teamhanko.github.io/hanko/).

## License
The hanko backend ist licensed under the [AGPL-3.0](LICENSE).
