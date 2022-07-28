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
  ## private ##
  #
  # Configuration for the private API.
  #
  private:
    ## address ##
    #
    # The address the private API will listen and handle requests on.
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
service:
  ## name ##
  #
  # The name of the service. This value will be used in the subject header of emails.
  #
  name: "Example Project"
## secrets ##
#
# Configures secrets used for signing. The secrets can be rotated by adding a new secret to the top of the list.
#
secrets:
  ## keys ##
  #
  # A secret that is used to sign and verify session JWTs. The first item is used for signing. The whole list is used for verifying session JWTs.
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
  ## enable_auth_token ##
  #
  # The JWT will be transmitted via the X-Auth-Token header. Enable during cross-domain operations.
  #
  enable_auth_token: false
password:
  ## enabled ##
  #
  # Enables or disables passwords for all users.
  #
  # Default value: false
  #
  enabled: false
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
    from_address: "CHANGE-ME"
    ## from_name ##
    #
    # The sender name of emails sent to users.
    #
    from_name: "CHANGE-ME"
  ## smtp ##
  #
  # SMTP server config to send emails.
  #
  smtp:
    host: "CHANGE-ME"
    ## port ##
    #
    # TODO:
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
    # The origin for which WebAuthn credentials will be accepted by the server. Must include the protocol and can only be the effective domain,
    # or a registrable domain suffix of the effective domain, as specified in the id. Except for localhost, the protocol must always be https for WebAuthn to work.
    #
    # Example:
    # - http://localhost
    # - https://example.com
    # - https://subdomain.example.com
    #
    origin: "http://localhost"
```

## Explanation

### Web Authentication

For most use cases, you just need the domain of your website that Hanko will be used with. Set `id` to the domain and set `origin` to the domain but include the protocol.

#### Examples

When you have a website hosted at `example.com` and you want to add a login to it that will be available
at `https://example.com/login`, the WebAuthn config would look like this:

```yaml
webauthn:
  relying_party:
    id: "example.com"
    display_name: "Example Project"
    origin: "https://example.com"
```

If the login should be available at `https://login.example.com` instead, then the WebAuthn config would look like this:

```yaml
webauthn:
  relying_party:
    id: "login.example.com"
    display_name: "Example Project"
    origin: "https://login.example.com"
```

Given the above scenario, you still may want to bind your users WebAuthn credentials to `example.com` if you plan to add other services on other subdomains later that should be able to use existing credentials. Another reason can be if you want to have the option to move your login from `https://login.example.com` to `https://example.com/login` at some point. Then the WebAuthn config would look like this:

```yaml
webauthn:
  relying_party:
    id: "example.com"
    display_name: "Example Project"
    origin: "https://login.example.com"
```

> **Note**: Currently, only a single origin is supported. We plan to add support for a list of origins at some point.
