# Manually testing custom claims (SAML + OIDC)

This walks through exercising the full custom-claims feature set against the
Keycloak-backed setup in `quickstart.keycloak.yaml` / `config-keycloak.yaml`. It
assumes you're in `deploy/docker-compose/`.

## Start the stack

```
docker compose -f quickstart.yaml -f quickstart.keycloak.yaml up -d
```

Keycloak imports three realms automatically (`uni-a`, `uni-b`, `acme-corp` - see
`keycloak-import/README.md`), so no manual Keycloak setup is needed.

**Known startup race:** Keycloak's JVM can take longer to bind its port than
Hanko waits at boot, so the initial SAML provider sync sometimes fails with
`connection refused`. If `docker compose logs hanko | grep saml` shows a
`WARNING: Failed to sync provider`, just run:

```
docker compose -f quickstart.yaml -f quickstart.keycloak.yaml restart hanko
```

and confirm the retry logs `Successfully synced provider` for both `uni-a` and
`uni-b`.

Credentials for every test user: username `alice`, password `alice123`
(same email, `alice@example.com`, in all three realms - that's what makes
Hanko link them into one user instead of creating three).

## Scenario 1 - SAML login via uni-a (all claim types)

1. Open `http://localhost:8888`, enter an email ending in `@uni-a.example.com`
   (any local part - only the domain drives the SAML redirect).
2. Log in as `alice` in the `uni-a` Keycloak realm.
3. Check the result:
   ```
   curl http://localhost:8001/users/<user-id>/custom_claims
   ```
   Expect:
   ```json
   {
     "matriculation_number": "urn:schac:personalUniqueCode:int:studentID:de:uni-a.example.com:12345",
     "affiliation": ["student@uni-a.example.com", "staff@uni-a.example.com"]
   }
   ```
   Two things worth noting, both intentional:
   - These are the **real** DFN-AAI/SCHAC attributes (`schacPersonalUniqueCode`,
     `eduPersonScopedAffiliation`), and their values are the raw structured
     strings the attributes actually carry - Hanko renames but does not
     reformat attribute values, so the full URN/`role@domain` shape comes
     through as-is.
   - `account_status` never appears. `uni-a` maps it from `schacUserStatus`,
     whose value (`urn:schac:userStatus:...`) can never parse as a boolean.
     Check the hanko logs for the resulting warning instead of a value:
     ```
     docker compose -f quickstart.yaml -f quickstart.keycloak.yaml logs hanko | grep account_status
     ```
     This demonstrates that a mistyped/mismatched claim mapping never fails a
     login - it's just skipped with a warning.

## Scenario 2 - source-scoped clear semantics (uni-b)

This is the core multi-connection behavior the feature exists for: a claim can
only be cleared by the same connection that last set it.

1. Log out, then log in again with an email ending in `@uni-b.example.com`,
   as `alice` in the `uni-b` realm. This links to the *same* Hanko user
   (matching email), it doesn't create a new one.
2. Re-check `custom_claims` - `matriculation_number`/`affiliation` are
   **unchanged**. `uni-b` declares both as managed claims but has no mapper
   providing values for either, so they're reported absent - and since the
   source on file is `uni-a`, not `uni-b`, an unrelated connection's silence
   does not clear them.
3. In the Keycloak admin console, `uni-a` realm -> Users -> `alice` -> remove
   both `affiliation` attribute values -> Save.
4. Log in via `uni-a` once more (same domain as step 1). Now `affiliation`
   **does** clear - the owning connection reporting a real absence is treated
   as a genuine signal, unlike an unrelated connection's silence in step 2.
5. Restore `alice`'s `affiliation` values in `uni-a` afterward if you want to
   repeat this test or move on to a clean state.

## Scenario 3 - OIDC custom provider (acme-corp)

Exercises the second, separate code path for resolving custom claims (OIDC/OAuth2
connections, as opposed to SAML), and proves the same claim mechanism produces
genuinely typed JSON values (not just strings) when the source is JSON-native.

1. On the login screen, click "Login with Acme Corp (Keycloak OIDC)".
2. Log in as `alice` in the `acme-corp` realm.
3. Check `custom_claims` again - it now also includes:
   ```json
   {
     "email_verified": true,
     "employee_id": 42
   }
   ```
   Both are genuine JSON `boolean`/`number` values (verify via
   `curl ... | python3 -m json.tool` - no quotes around `true`/`42`), sourced
   from the OIDC ID token, tagged `"source": "third_party:custom_acme_corp"`.

## Scenario 4 - Admin PATCH survives unrelated connection logins

```
curl -X PATCH http://localhost:8001/users/<user-id>/custom_claims \
  -H "Content-Type: application/json" \
  -d '{"affiliation": ["admin-set-value"]}'
```

Log in via any connection that doesn't manage `affiliation` (e.g. `acme-corp`
again). Check `custom_claims` - `affiliation` is still `["admin-set-value"]`,
tagged `"source": "admin"`. Admin-set values are protected by the same
source-scoped logic as connection-set ones.

## Scenario 5 - overwrite semantics (both uni-a and uni-b provide home_organization)

Scenario 2 only exercises the *absence/clear* side of refresh semantics (uni-b
never provides a value at all). This scenario exercises the other side: two
connections that both genuinely provide a value for the same claim - does the
write correctly overwrite unconditionally (last-one-wins, no source-ownership
check), regardless of who "owns" the claim?

`home_organization` (`schacHomeOrganization`) is used for this specifically so
it doesn't interfere with Scenario 2 - both `uni-a` and `uni-b` map it, each to
their own real, distinct domain value.

1. Log in via `uni-a`. Check `custom_claims` - `home_organization` is
   `"uni-a.example.com"`, `source: "saml:...uni-a"`.
2. Log in via `uni-b`. Check again - `home_organization` is now
   `"uni-b.example.com"`, `source` flipped to `uni-b`. No absence/clear logic
   is involved here at all; this is a plain overwrite, since both connections
   always provide a real value.

## Scenario 6 - user.update webhook fires with custom_claims

```
docker compose -f quickstart.yaml -f quickstart.keycloak.yaml logs webhook-receiver
```

After any login that changes a claim (e.g. Scenario 5's `uni-b` step), this
should show a `POST /hook` with a JSON body `{"token": "<JWT>", "event":
"user.update"}`. The payload isn't HMAC-signed with a separate secret - decode
the JWT (no need to verify its signature for this manual check) to see the
event data, which should include a `custom_claims` object matching whatever
`GET .../custom_claims` currently returns:

```
python3 -c "
import base64, json
token = '<paste the token value here>'
payload = token.split('.')[1]
payload += '=' * (-len(payload) % 4)
print(json.dumps(json.loads(base64.urlsafe_b64decode(payload)), indent=2))
"
```

This is a smoke test for the webhook trigger, not a full webhook consumer -
`webhook-receiver` (a generic `mendhak/http-https-echo` container) just logs
whatever it receives.

**Two real bugs were found and fixed via this exact test**, both in
`backend/`, unrelated to Keycloak/config:
- `handler/public_router.go`'s `/saml/callback` route was missing
  `webhookMiddleware` (present on the equivalent `/thirdparty/callback`
  routes), so `TriggerWebhooks` silently failed for every SAML login - the
  webhook was configured correctly and simply never fired.
- `thirdparty/linking.go`'s `LinkAccount` never refreshed `result.User`'s
  eager-loaded `CustomClaims` after `applyCustomClaims` wrote a change, so the
  webhook payload (built from that `User`) carried one-write-stale claim data
  even once the middleware fix made it fire at all.

## Notes for reviewers

- `matriculation_number`/`affiliation`/`account_status` are modeled on real
  DFN-AAI/SCHAC attributes (see `config-keycloak.yaml` comments for the exact
  OIDs and why each was chosen this way, including the deliberate
  boolean-coercion-failure case).
- `uni-b` deliberately declares `matriculation_number`/`affiliation` as
  managed without ever providing values (proving silence is harmless in
  Scenario 2), while also genuinely providing `home_organization` (proving
  overwrite in Scenario 5) - both are real, distinct behaviors of the same
  connection.
- Hanko's own database (`user_custom_claims`, users, identities) is **not**
  captured by the Keycloak realm export and has no persistent volume - it
  starts empty on every fresh `docker compose up`. Only the Keycloak-side
  state (realms/clients/mappers/users) is reproducible via
  `keycloak-import/*.json`; the claim data itself is generated by actually
  walking through the scenarios above.
