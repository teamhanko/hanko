# Config options

## All available config keys
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
    secure: true
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

## Explanation

TBD
