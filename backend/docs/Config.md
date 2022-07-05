# Hanko backend config

All config parameters with their default and allowed values are documented here. For some parameters there is an extra
section with more detailed instructions below.

## All available config options

```yaml
## Hanko Service configuration ##
#

server:
  ## public ##
  #
  # Configuration for the public API
  #
  public:
    ## address ##
    #
    # The address the public API will listen and handle requests on.
    #
    address: ":8000"
    ## cors ##
    #
    # Configures Cross Origin Resource Sharing for public endpoints.
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
  private:
    ## address ##
    #
    # The address the private API will listen and handle requests on.
    #
    address: ":8001"
## database ##
#
# This configures the backend where to persist data.
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
  # This configures the backend which database will be used.
  #
  # One of:
  # - cockroach
  # - mariadb
  # - mysql
  # - postgres
  #
  dialect: "postgres"
  user: "hanko"
  password: "hanko"
  database: "hanko"
service:
  ## name ##
  #
  # This is the name of the service. This value will be used in the subject header of emails
  #
  name: "Hanko Authentication Service"
## secrets ##
#
# The secrets section configures secrets used for signing. The secrets can be rotated by adding a new secret to the top of the list.
#
secrets:
  ## keys ##
  #
  # A secret that is used to sign and verify session JWTs. The first item is used for signing. The whole list is used for verifying session JWTs.
  #
  keys:
    - "change-me"
session:
  ## lifespan ##
  #
  # Sets how long a session JWT is valid.
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
    # Sets on which domain the cookie will be bound. Does not work cross domain, but only for subdomains.
    #
    domain: ""
    ## http_only ##
    #
    # Sets whether the cookie is a http only cookie or can be read by javascript.
    #
    # Default value: true
    #
    http_only: true
    ## same_site ##
    #
    # Configures the same site attribute of the session cookie.
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
password:
  ## enabled ##
  #
  # Configures whether password are enabled or not.
  #
  # Default value: false
  #
  enabled: false
passcode:
  ## ttl ##
  #
  # Sets how long a passcode is valid. Value is in seconds.
  #
  # Default value: 300
  #
  ttl: 0
  email:
    ## from_address ##
    #
    # Configures the sender of emails sent to the users.
    #
    from_address: ""
    ## from_name ##
    #
    # Configures the sender name of emails sent to the users.
    #
    from_name: ""
  ## smtp ##
  #
  # Configures the backend which smtp server will be used to sent emails.
  #
  smtp:
    host: ""
    ## port ##
    #
    # TODO:
    #
    # Default: 465
    #
    port: ""
    user: ""
    password: ""
## webauthn ##
#
# Configures web authentication
#
webauthn:
  ## timeout ##
  #
  # Configures how long a web authentication request is valid and the user can confirm it. Value is in milliseconds
  #
  # Default: 60000
  #
  timeout: 0
  relying_party:
    ## id ##
    #
    # ID sets the host on which web authentication can be used.
    #
    # Examples:
    # - localhost
    # - example.com
    # - subdomain.example.com
    #
    id: "localhost"
    ## display_name ##
    #
    # Sets the name which the web authentication authenticator will show during the ceremony
    #
    display_name: ""
    ## origin ##
    #
    # Sets the origin of which web authentication can be used.
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

For most use cases, just add the host name of your website / app on which you want to use WebaAuthn as the id and set the origin by including the scheme in the config.

#### Example

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
