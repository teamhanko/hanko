# Docker Compose setups

Local Hanko stacks for different purposes. All commands below assume you're in
this directory.

## Quickstart

```
docker compose -f quickstart.yaml up -d
```

The standard local dev stack: `hanko` (+ `hanko-migrate`), `postgresd`,
`elements` (the frontend web components), `quickstart` (the example app, port
8888), `mailslurper` (a local SMTP catcher with a web UI on port 8080).
`config.yaml` is its config file.

Variants of this same stack:

- **`quickstart-with-redis.yaml`** - adds a `redis` service, for testing
  Hanko against Redis-backed session storage/rate limiting instead of the
  default.
- **`quickstart.debug.yaml`** - builds `hanko`/`elements` from
  `Dockerfile.debug` with `SYS_PTRACE`/`apparmor=unconfined` and exposes port
  `40000`, for attaching a remote debugger.
- **`quickstart.e2e.yaml`** - adds an `e2e` service that seeds the database
  for the end-to-end test suite (`../../e2e`). **Very old and not part of the
  automatic CI pipeline** (`.github/workflows/e2e.yml` is `workflow_dispatch`
  only, never triggered on push/PR) - either remove this manifest and the e2e
  suite, or update both so they're actually maintained and run in CI.
- **`quickstart.keycloak.yaml`** - see [Keycloak](#keycloak-saml--oidc-testing)
  below.

## Config variants

- **`config.yaml`** - the default config for `quickstart.yaml`.
- **`config-disable-signup.yaml`** - signup disabled, for testing login-only
  flows.
- **`config-rate-limiting.yaml`** - rate limiting enabled (used by
  `quickstart-with-redis.yaml`).
- **`config-keycloak.yaml`** - see below.

## Todo example apps

```
docker compose -f base.yaml -f todo-<framework>.yaml -p "hanko-todo-<framework>" up --build
```

`base.yaml` provides the shared backend (`hanko`, `postgresd`, `elements`,
`mailslurper`); each `todo-<framework>.yaml` (`react`, `vue`, `angular`,
`nextjs`) adds that framework's example frontend + its small Express backend.
Documented in `../../frontend/examples/README.md`.

## Keycloak (SAML / OIDC testing)

```
docker compose -f quickstart.yaml -f quickstart.keycloak.yaml up -d
```

An override file (merges on top of `quickstart.yaml`, which stays untouched)
that adds a local Keycloak instance - it acts as both a SAML 2.0 IdP and a
full OIDC provider, so it covers both connection types Hanko supports. Used to
manually test SAML/OIDC login and the custom-claims feature end to end,
without needing a real university/enterprise IdP.

- `keycloak` + `keycloak-db-init` run only when this override file is
  included - a plain `docker compose -f quickstart.yaml up` never touches
  Keycloak or its database.
- Keycloak's realms (`uni-a`, `uni-b`, `acme-corp` - modeling two SAML
  federation partners and one generic OIDC enterprise provider) are
  pre-configured and auto-imported from `keycloak-import/*.json`, so no
  manual admin-console setup is needed. See `keycloak-import/README.md` for
  how to update/re-export those files after changing the Keycloak setup.
- `hanko`/`hanko-migrate` use **`config-keycloak.yaml`** instead of the
  default `config.yaml` in this override - a full copy with the
  `saml`/`third_party`/`custom_claims` sections added (and comments
  explaining the specific DFN-AAI/SCHAC attributes and OIDC claims modeled).

For a full step-by-step walkthrough of what to click and what to expect (all
claim types, multi-connection refresh semantics, OIDC, Admin API), see
**[TESTING.md](./TESTING.md)**.
