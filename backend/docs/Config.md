# Hanko backend config

The Hanko backend can be configured using a `yaml` configuration file or using environment variables.
Environment variables have higher precedence than configuration via file (i.e. if provided, they overwrite the values
given in the file - multivalued options, like arrays, are also _not_ merged but overwritten entirely).

The schema for the configuration file is given below. To set equivalent environment variables, join keys by `_`
(underscore) and uppercase the keys, i.e. for `server.webauthn.relying_party.origins`
use:

```shell
export SERVER_WEBAUTHN_RELYING_PARTY_ORIGINS="https://hanko.io,android:apk-key-hash:nLSu7w..."
```


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
      ## allow_origins ##
      #
      # A list of allowed origins that may access the public endpoints.
      #
      allow_origins:
        - "https://hanko.io"
        - "https://example.com"
      ## unsafe_wildcard_origin_allowed ##
      #
      # If allow_origins contains a wildcard '*' origin, this flag must explicitly be set to 'true'.
      # Using wildcard '*' origins is insecure and potentially leads to cross-origin attacks.
      #
      unsafe_wildcard_origin_allowed: false
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
    ## name ##
    #
    # Sets the name of the cookie.
    #
    # Default value: hanko
    #
    name: true
  ## enable_auth_token_header ##
  #
  # The JWT will be transmitted via the X-Auth-Token header. Enable during cross-domain operations.
  #
  enable_auth_token_header: false
  ## audience ##
  #
  # Audience optional []string containing strings which get put into the aud claim. If not set default to Webauthn.RelyingParty.Id config parameter.
  #
  audience:
  ## issuer ##
  #
  #  optional string to be used in the jwt iss claim.
  #
  issuer:
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
  ## user_verification ##
  #
  # Describes your requirements regarding local authorization with an authenticator through
  # various authorization gesture modalities; for example, through a touch plus pin code,
  # password entry, or biometric recognition.
  #
  # Must be one of "required", "preferred" or "discouraged".
  #
  # The setting applies to both WebAuthn registration and authentication ceremonies.
  #
  # Default: preferred
  #
  user_verification: preferred
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
    ## origins ##
    #
    # A list of origins for which WebAuthn credentials will be accepted by the server. Must include the protocol and can only be the effective domain,
    # or a registrable domain suffix of the effective domain, as specified in the id. Except for localhost, the protocol must always be https for WebAuthn to work.
    # Ip Addresses will not work.
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
    # When to reset the token interval
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
    # How many operations can occur in the given interval
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
emails:
  ## require_verification
  #
  # If turned on (true), email addresses must be verified to be assigned to a user. During account creation the user has
  # to verify the email address before a JWT can be issued and further email addresses can only be added after they have
  # been verified. If require_verification is turned off, the mentioned verification steps aren't necessary (and will be
  # skipped by hanko-elements), so the user can create an account without email verification and further email addresses
  # can be added directly.
  #
  # Default: true
  #
  require_verification: true
  ## max_num_of_addresses
  #
  # How many email addresses can be added to a user account
  #
  # Default: 5
  #
  max_num_of_addresses: 5
## third_party ##
#
# Configures third party providers
#
third_party:
  ## redirect_url
  #
  # Required if any providers are enabled.
  # The URL the third party provider redirects to with an authorization code. Must consist of the base URL of
  # your running hanko backend instance and the callback endpoint, i.e. <YOUR_BACKEND_INSTANCE>/thirdparty/callback.
  #
  redirect_url: "CHANGE_ME"
  ## allowed_redirect_urls
  #
  # Required if any providers are enabled.
  # List of URLS the backend is allowed to redirect to after third party sign-in was successful.
  # (see also the 'redirect_to' parameter for the third party auth initialization endpoint
  # - https://docs.hanko.io/api/public#tag/Third-Party/operation/thirdPartyAuth)
  #
  # Supports wildcard matching through globbing. e.g. https://*.example.com will allow https://foo.example.com and https://bar.example.com to be accepted.
  # Globbing is also supported for paths, e.g. https://foo.example.com/* will match https://foo.example.com/page1 and https://foo.example.com/page2.
  # A double asterisk (`**`) acts as a "super"-wildcard/match-all.
  #
  # More on globbing: https://pkg.go.dev/github.com/gobwas/glob#Compile
  #
  # NOTE: URLs in the list MUST NOT have trailing slash
  #
  # Examples:
  # - http://localhost:8888
  #
  allowed_redirect_urls:
    - "CHANGE_ME"
  ##
  #
  # Required if any providers are enabled. URL the backend redirects to if an error occurs during third party sign-in. Errors are provided
  # as 'error' and 'error_description' query params in the redirect location URL.
  #
  # When using the Hanko web components it should be the URL of the page that embeds the web component such that
  # errors can be processed properly by the web component.
  #
  # You do not have to add this URL to the 'allowed_redirect_urls', it is automatically included when validating
  # redirect URLs.
  #
  # NOTE: MUST NOT have trailing slash
  #
  # Example:
  # - http://localhost:8888/error
  #
  error_redirect_url: "CHANGE_ME"
  ##
  #
  # The third party provider configurations. Unknown providers will be ignored.
  #
  providers:
    ##
    #
    # The Apple provider configuration
    #
    apple:
      ##
      #
      # Enable or disable the Google provider.
      #
      # Default: false
      #
      enabled: false
      ##
      #
      # The client ID (Services ID) of your Apple credentials.
      # See: https://docs.hanko.io/guides/authentication-methods/oauth/apple
      #
      # Required if provider is enabled.
      #
      #
      client_id: "CHANGE_ME"
      ##
      #
      # The generated secret of your Apple credentials.
      # Valid for max. 6 months. Must be regenerated before expiration.
      #  https://docs.hanko.io/guides/authentication-methods/oauth/apple
      #
      # Required if provider is enabled.
      #
      secret: "CHANGE_ME"
    ##
    #
    # The Google provider configuration
    #
    google:
      ##
      #
      # Enable or disable the Google provider.
      #
      # Default: false
      #
      enabled: false
      ##
      #
      # The client ID of your Google OAuth credentials.
      # See: https://docs.hanko.io/guides/authentication-methods/oauth/google
      #
      # Required if provider is enabled.
      #
      client_id: "CHANGE_ME"
      ##
      #
      # The secret of your Google OAuth credentials
      # See: https://docs.hanko.io/guides/authentication-methods/oauth/google
      #
      # Required if provider is enabled.
      #
      secret: "CHANGE_ME"
    ##
    #
    # The GitHub provider configuration
    #
    github:
      ##
      #
      # Enable or disable the GitHub provider.
      #
      # Default: false
      #
      enabled: false
      ##
      #
      # The client ID of your GitHub OAuth credentials.
      # See: https://docs.hanko.io/guides/authentication-methods/oauth/github
      #
      # Required if provider is enabled.
      #
      client_id: "CHANGE_ME"
      ##
      #
      # The secret of your GitHub OAuth credentials.
      # See: https://docs.hanko.io/guides/authentication-methods/oauth/github
      #
      # Required if provider is enabled.
      #
      secret: "CHANGE_ME"
log:
  ## log_health_and_metrics
  #
  # If turned on (true) also logs invocations of the /health and /metrics endpoints
  #
  # Default: false
  #
  log_health_and_metrics: false
account:
  ## allow_deletion
  #
  # The user account can be deleted by the current user when turned on, otherwise the corresponding API route will not
  # be available.
  #
  # Default: false
  #
  allow_deletion: false
  ## allow_signup
  #
  # Users are able to sign up new accounts.
  #
  # Default: true
  #
  allow_signup: true
```
