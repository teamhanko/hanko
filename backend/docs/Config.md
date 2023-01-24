# Hanko backend config

All config parameters with their defaults and allowed values are documented here. For some parameters there is an extra
section with more detailed instructions below.

## All available config options

```yaml
## Hanko Service configuration ##
#

server:
  ## public ##
  #
  # Configuration for the public API.
  #
  public:
    ## address ##
    #
    # The address the public API will listen and handle requests on.
    #
    address: ":8000"
    ## cors ##
    #
    # Cross Origin Resource Sharing for public endpoints.
    #
    cors:
      ## enabled ##
      #
      # Sets whether cors is enabled or not.
      #
      # Default value: false
      #
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
  ## admin ##
  #
  # Configuration for the admin API.
  #
  admin:
    ## address ##
    #
    # The address the admin API will listen and handle requests on.
    #
    address: ":8001"
## database ##
#
# Configures the backend where to persist data.
#
database:
  host: "localhost"
  ## port ##
  #
  # Default value: 5432
  #
  port: "5432"
  ## dialect ##
  #
  # Which database will be used.
  #
  # One of:
  # - cockroach
  # - mariadb
  # - mysql
  # - postgres
  #
  dialect: "postgres"
  user: "CHANGE-ME"
  password: "CHANGE-ME"
  database: "CHANGE-ME"
  ## url ##
  #
  # Instead of using the individual fields above this field can be used.
  # When this field is set, it will be used and the fields above have no effect.
  #
  # Url schema: `dialect://username:password@host:port/database`
  #
  # Examples:
  # - postgres://hanko:hanko@localhost:5432/hanko
  #
  url: "CHANGE-ME"
service:
  ## name ##
  #
  # The name of the service. This value will be used in the subject header of emails.
  #
  name: "Example Project"
## secrets ##
#
# Configures secrets used for en-/decrypting JWKs.
#
secrets:
  ## keys ##
  #
  # Keys secrets are used to en- and decrypt the JWKs which get used to sign the JWTs.
  # For every key a JWK is generated, encrypted with the key and persisted in the database.
  #
  # You can use this list for key rotation: add a new key to the beginning of the list and the corresponding
  # JWK will then be used for signing JWTs. All tokens signed with the previous JWK(s) will still
  # be valid until they expire. Removing a key from the list does not remove the corresponding
  # database record. If you remove a key, you also have to remove the database record, otherwise
  # application startup will fail.
  #
  # Each key must be at least 16 characters long.
  #
  keys:
    - "CHANGE-ME"
session:
  ## lifespan ##
  #
  # How long a session JWT is valid.
  #
  # Default value: 1h
  #
  # Examples:
  # - 1h
  # - 10m
  # - 720h
  # - 15h115m
  #
  lifespan: "1h"
  cookie:
    ## domain ##
    #
    # The domain the cookie will be bound to. Works for subdomains, but not cross-domain.
    #
    domain: "CHANGE-ME"
    ## http_only ##
    #
    # HTTP-only cookies or accessible by javascript.
    #
    # Default value: true
    #
    http_only: true
    ## same_site ##
    #
    # Same-site attribute of the session cookie.
    #
    # Default value: strict
    #
    # One of:
    # - strict
    # - lax
    # - none
    #
    same_site: "strict"
    ## secure ##
    #
    # Sets whether the cookie can only be read on secure sites.
    #
    # Default value: true
    #
    secure: true
  ## enable_auth_token_header ##
  #
  # The JWT will be transmitted via the X-Auth-Token header. Enable during cross-domain operations.
  #
  enable_auth_token_header: false
password:
  ## enabled ##
  #
  # Enables or disables passwords for all users.
  #
  # Default value: false
  #
  enabled: false
  ## min_password_length ##
  #
  # Sets the minimum password length.
  #
  # Default value: 8
  #
  min_password_length: 8
passcode:
  ## ttl ##
  #
  # How long a passcode is valid. Value is in seconds.
  #
  # Default value: 300
  #
  ttl: 300
  email:
    ## from_address ##
    #
    # The sender of emails sent to users.
    #
    # Default value: passcode@hanko.io
    #
    from_address: passcode@hanko.io"
    ## from_name ##
    #
    # The sender name of emails sent to users.
    #
    # Default value: Hanko
    #
    from_name: "Hanko"
  ## smtp ##
  #
  # SMTP server config to send emails.
  #
  smtp:
    host: "CHANGE-ME"
    ## port ##
    #
    # Default: 465
    #
    port: ""
    user: "CHANGE-ME"
    password: "CHANGE-ME"
## webauthn ##
#
# Configures Web Authentication (WebAuthn).
#
webauthn:
  ## timeout ##
  #
  # How long a WebAuthn request is valid and the user can confirm it. Value is in milliseconds.
  #
  # Default: 60000
  #
  timeout: 60000
  relying_party:
    ## id ##
    #
    # The effective domain the WebAuthn credentials will be bound to.
    #
    # Examples:
    # - localhost
    # - example.com
    # - subdomain.example.com
    #
    id: "localhost"
    ## display_name ##
    #
    # The service's name that some WebAuthn Authenticators will display to the user during registration and authentication ceremonies.
    #
    # Examples:
    # - Example Project
    # - Hanko GmbH
    # - Acme, Inc.
    #
    display_name: ""
    ## origin ##
    #
    # DEPRECATED: use "origins" instead
    #
    # The origin for which WebAuthn credentials will be accepted by the server. Must include the protocol and can only be the effective domain,
    # or a registrable domain suffix of the effective domain, as specified in the id. Except for localhost, the protocol must always be https for WebAuthn to work.
    #
    # Example:
    # - http://localhost
    # - https://example.com
    # - https://subdomain.example.com
    #
    origin: "http://localhost"
    ## origins ##
    #
    # A list of origins for which WebAuthn credentials will be accepted by the server. Must include the protocol and can only be the effective domain,
    # or a registrable domain suffix of the effective domain, as specified in the id. Except for localhost, the protocol must always be https for WebAuthn to work.
    #
    # For an Android app the origin must be the base64 url encoded SHA256 fingerprint of the signing certificate.
    #
    # Example:
    # - android:apk-key-hash:nLSu7wVTbnMOxLgC52f2faTnvCbXQrUn_wF9aCrr-l0
    # - https://login.example.com
    #
    origins:
      - "android:apk-key-hash:nLSu7wVTbnMOxLgC52f2faTnv..."
      - "https://login.example.com"
## audit_log ##
#
# Configures audit logging
#
audit_log:
  console_output:
    ## enabled ##
    #
    # Sets whether the output to console is enabled or disabled.
    #
    # Default: true
    #
    enabled: true
    ## output ##
    #
    # The output stream which audit logs are sent to.
    #
    # Possible values:
    # - stdout
    # - stderr
    #
    # Default: stdout
    #
    output: "stdout"
  storage:
    ## enabled ##
    #
    # Sets whether the audit logs are persisted in the database or not.
    #
    # Default: false
    #
    enabled: false
rate_limiter:
  ## enabled ##
  #
  # Sets whether the rate limiting is enabled or disabled
  #
  # Default: true
  #
  enabled: true
  ## store ##
  #
  # Sets the store for the rate limiter. When you have multiple instances of Hanko running, it is recommended to use 
  # the "redis" store else your instances have their own states.
  #
  # One of:
  # - in_memory
  # - redis
  #
  # Default: in_memory
  #
  store: "in_memory"
  ## password_limits
  #
  # rate limits specific to the password/login endpoint
  #
  password_limits:
    ## tokens
    #
    # How many operations can occur in the given interval
    #
    # Default: 5
    tokens: 5
    ## interval
    #
    # When to reset the token interval?
    #
    # Default: 1m
    #
    interval: 1m
  ## password_limits
  #
  # rate limits specific to the passcode/init endpoint
  #
  passcode_limits:
    ## tokens
    #
    # How many operations can occur in the given interval?
    #
    # Default: 3
    tokens: 3
    ## interval
    #
    # When to reset the token interval
    #
    # Default: 1m
    #
    interval: 1m
  ## redis_config
  #
  # If you specify redis as backend you have to specify these values
  #
  redis_config:
    ## address
    #
    # Address of your redis instance in the form of host[:port][/database]
    #
    address: "CHANGE-ME"
    ## password
    #
    # The password of the redis instance
    #
    password: "CHANGE_ME"
```
