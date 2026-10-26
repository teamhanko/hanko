# Threat model

The reporting policy is `SECURITY.md` in the repository root: report privately to security@hanko.io.

## What Hanko does and where untrusted input enters

Hanko is an open-source authentication and user-management server written in Go (`backend/`). It provides
passkey/WebAuthn, passcode (email OTP), password, TOTP and security-key MFA, social login (OAuth2/OIDC), SAML SSO,
sessions issued as signed JWTs, webhooks, organizations and roles, and optional multitenancy. Applications embed the
`hanko-elements` web components and the `frontend-sdk` (`frontend/`) and verify the JWTs Hanko issues.

The backend exposes three HTTP servers:

- **Public API** (`:8000`, `handler/public_router.go`): fully untrusted, internet-facing. The login, registration,
  profile and token-exchange flows (`flow_api/`, built on the `flowpilot/` state machine), `/me`, `/logout`,
  `/sessions/validate`, `/thirdparty/callback`, `/saml/*`, `/.well-known/jwks.json` and
  `/organizations/roles/check`. In multitenancy mode every route is prefixed with `/:tenant_id`.
- **Admin API** (`:8001`, `handler/admin_router.go`): has no authentication of its own. The documentation requires
  operators to put it behind their own access control. Callers are trusted.
- **Management API** (`:8002`, `handler/management_router.go`): tenant and SAML-provider CRUD, metrics. Trusted, like
  the admin API.

Untrusted input therefore includes:

- every request body, header, cookie, query parameter and path parameter on the public API, including the flow
  action/input payloads and the flow ID;
- WebAuthn attestation and assertion responses, client data and authenticator data;
- OAuth2/OIDC callback parameters (`state`, `code`, errors) and the ID tokens and userinfo returned by identity
  providers, including custom OIDC providers;
- SAML responses and assertions posted to the callback, and IdP metadata (attacker-controlled if an attacker
  operates an IdP a tenant has configured);
- session JWTs and auth-token headers presented by clients;
- `X-Forwarded-For` and similar headers, as interpreted by the configured IP extractor (`server.ip`);
- email addresses and usernames, which flow into outbound emails and templates.

Configuration files, environment variables, the database and anything sent to the admin or management API are
trusted.

## Security goals

1. Nobody can authenticate as, or obtain a valid session for, a user without that user's credentials, or alter a
   user's identities, credentials, emails or MFA settings without satisfying the required verification.
2. Every factor the configuration requires (email verification, MFA, passkey, password) is actually enforced; a
   flow cannot be steered into a later state while skipping a required step.
3. Sessions are bound to their user and tenant, can be revoked, and expire as configured. JWTs are signed with the
   tenant's keys and cannot be forged, replayed across tenants or used after revocation where revocation is enabled.
4. In multitenancy mode, tenants are fully isolated: users, sessions, keys, credentials, SAML providers,
   organizations and configuration of one tenant are never readable or usable from another.
5. Organization membership and role checks cannot be satisfied by a user who does not hold the role.
6. Secrets (passcodes, password hashes, TOTP secrets, private JWKs, flow state, webhook secrets, provider client
   secrets) are not disclosed to the public API.
7. Rate limits on passcodes, passwords, OTP and token exchange cannot be bypassed in a way that enables practical
   brute force.

## Components to prioritize

Highest priority:

- `flow_api/` and `flowpilot/`: the login, registration, profile and token-exchange state machines. Look for state
  confusion, skipped or replayed actions, actions accepted in the wrong state, missing checks that a step was
  completed (email verified, MFA passed), flow-ID fixation or reuse across users or tenants, and races between
  concurrent requests on the same flow (`flow_api/flow_locker`).
- `flow_api/services/`: passcode, password, TOTP and WebAuthn verification, and device trust.
- `session/` and `middleware/`: JWT creation and validation, cookie handling, session revocation, the tenant
  middleware, and IP extraction.
- `thirdparty/`: OAuth2/OIDC providers, `state` handling, account linking (`linking.go`) and custom claims. Account
  takeover through linking an unverified provider email to an existing account is critical.
- `saml/` and `handler/saml.go`: assertion validation, signature checks, replay, audience and recipient checks,
  RelayState and redirect handling, IdP-initiated login.
- `crypto/`: passcode generation, JWK handling and AES-GCM encryption of stored secrets.
- `handler/organization_public.go`, `persistence/` queries scoped by tenant and organization: cross-tenant or
  cross-organization access.

Medium priority:

- `webhooks/`: outbound requests to configured URLs (SSRF is only relevant where a lower-privileged actor can
  influence the URL), signing, and leakage of data across tenants in webhook payloads.
- `mail/` and `template/`: header injection and template injection through user-controlled values.
- `rate_limiter/`.
- `frontend/` (`hanko-elements`, `frontend-sdk`): XSS or open redirects in the web components, and token handling in
  the SDK.

Lower priority:

- the admin and management APIs (trusted callers; only report issues reachable without admin access, or that let an
  admin of one tenant affect another tenant in multitenancy mode);
- `cmd/` (CLI tools run by operators), `json_schema/`, `docs/`, `deploy/`, `quickstart/`, `e2e/`.

## How to exercise the image

The checkout is at `/src`; the Go module is `/src/backend`, built to `/usr/local/bin/hanko`. Go modules are cached and
`GOPROXY=off`, so `go build` and `go test` work offline.

A local PostgreSQL is installed with role `hanko` / password `hanko` and databases `hanko` (schema migrated) and
`hanko_test`, matching `backend/config/config.yaml`. To run a server:

```sh
service postgresql start
cd /src/backend
hanko serve all --config config/config.yaml   # public :8000, admin :8001, management :8002
```

Copy and edit `config/config.yaml` to enable
features under test (multitenancy, SAML, third-party providers, webhooks, username login, etc.). No SMTP server runs
in the image; passcode emails are logged as delivery failures unless you set `email_delivery.enabled: false` or run a
local SMTP sink. Run `hanko migrate up --config <file>` after pointing a config at a fresh database.

`go test -short ./...` runs the unit tests. The database-backed suites (`handler`, `persistence`, `saml`,
`webhooks`, `cmd/user`, `crypto/jwk/local_db`) normally start PostgreSQL through dockertest, which needs Docker and is
not available in the scanner; they are skipped with `-short`. To exercise flows end to end, drive the running server
over HTTP with `curl` (the flow API is JSON: `POST /login` returns a flow with actions, and each action is posted back
to the URL it lists).

## Severity

- Critical: authentication bypass or account takeover of an arbitrary user over the public API without user
  interaction (e.g. skipping a required factor, forging a session JWT, linking a third-party or SAML identity to
  another user's account, passcode or password verification bypass); cross-tenant access to users, sessions or keys in
  multitenancy mode; remote code execution.
- High: account takeover that needs limited victim interaction; bypass of required MFA or email verification; session
  revocation that does not take effect; disclosure of secrets (private JWKs, passcodes, password hashes, TOTP secrets,
  client secrets) through the public API; SAML signature or replay bypass; practical brute force of passcodes or
  passwords through a rate-limit bypass; stored XSS in hanko-elements.
- Medium: user enumeration where the configuration says it should be prevented; reflected XSS or open redirect;
  disclosure of non-secret personal data of other users; privilege escalation within an organization; SSRF reachable
  by a non-admin.
- Low: denial of service that needs sustained traffic or large payloads; issues that require a misconfiguration the
  documentation warns against; information leaks with no practical impact; defense-in-depth hardening.

## Reports and exclusions

Include the Hanko commit, the relevant configuration (as a minimal YAML diff from `backend/config/config.yaml`), the
exact sequence of HTTP requests and responses, and why the result crosses one of the goals above.

Do not report:

- anything that requires access to the admin or management API, the database, the configuration or the signing keys,
  except cross-tenant effects in multitenancy mode;
- the admin or management API being unauthenticated: it is documented to require external access control;
- insecure behavior that only occurs with options explicitly documented as insecure (e.g. wildcard CORS with
  `unsafe_wildcard_origin_allowed`, `secure: false` cookies, trusting forwarded headers without a proxy);
- default secrets, keys or credentials in sample, test, `deploy/` or `quickstart/` configurations;
- vulnerabilities in dependencies that are not reachable through Hanko;
- missing security headers, verbose error messages or absent rate limits on endpoints with no security impact.
