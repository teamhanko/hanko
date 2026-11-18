# Proposal: self-service organizations through the Flow API

Status: draft for discussion. This document proposes API and data-model changes; none of the endpoints or configuration options below are implemented by this PR.

## Problem and scope

A static application can authenticate users with Hanko Elements but still needs an application backend to create organizations, invite users, and manage membership roles safely. The browser cannot hold an Admin API credential.

Provide these capabilities through Hanko's public API, authenticated with the user's ordinary Hanko session. Hanko remains the trusted backend. Organizations are groups inside a Hanko tenant; they are not Hanko tenants or Cloud console organizations.

The first use case is:

- Every newly registered user receives one personal organization and becomes its owner.
- Its display name is optional.
- Owners and admins invite members by email and distinguish pending invitations from joined members.
- A user can join additional organizations without losing their personal organization.
- Owners appoint admins and transfer ownership to an existing member.
- An application can build a custom UI or eventually embed an organization Element.

Personal-organization provisioning and self-service management must be opt-in tenant features. Existing installations and arbitrary application roles must retain their current semantics.

## Existing foundation

[PR #2893](https://github.com/teamhanko/hanko/pull/2893) introduced organizations, memberships, tenant-wide role definitions, organization-scoped role bindings, and a public role check.

Relevant implementation points:

- [Public routing](../handler/public_router.go) registers login, registration, profile, and token-exchange flows and `POST /organizations/roles/check`.
- [Public role checks](../handler/organization_public.go) resolve the session subject with `GetByPublicID` in the current tenant and check persisted membership and bindings.
- [Organization validation](../persistence/models/organization.go) currently requires a name.
- [Organization migration](../persistence/migrations/20260903120000_create_organizations.up.fizz) makes names unique within a tenant.

Use those persistence foundations rather than introducing a separate membership store. The existing generic role model does not itself enforce a single owner or a management hierarchy.

## API shape

Use flows for interactive mutations and paginated resource endpoints for reads. All paths below are proposed and relative to the existing tenant-aware public router.

| Method and path | Purpose |
| --- | --- |
| `GET /me/organizations` | List the caller's organizations for a switcher |
| `GET /organizations/{id}/members` | List members and their management roles |
| `GET /organizations/{id}/invitations` | List invitations and their lifecycle status |
| `POST /organization` | Initialize or advance organization management |
| `POST /organization-invitation` | Initialize or advance invitation acceptance |

Read endpoints require a valid session and organization-specific authorization; the invitation list is limited to owners/admins. Use existing pagination conventions, bounded page sizes, and minimal public DTOs. Do not expose internal user identifiers, private metadata, invitation secrets, or unrestricted account email collections. For inaccessible organizations, use a consistent not-found response.

An organization flow starts with an `organization_id` in its initialization input. Resolve it against the current tenant and session user. Subsequent action requests use the server-returned action URLs, inputs, and CSRF token, following the existing Flow API envelope. Organization context is immutable for that flow; switching organizations starts a new flow.

The `organization_manage` state contains organization metadata, the caller's management role, and permitted actions. Member and invitation tables load through the read endpoints instead of embedding unbounded collections in every state.

Suggested actions:

- `rename_organization`
- `invite_member`
- `resend_invitation`
- `revoke_invitation`
- `change_member_role`
- `remove_member`
- `transfer_ownership`
- `leave_organization`

Simple actions return to the management state after one mutation. Do not require a multi-screen wizard for renaming an organization. Owner transfer and similarly sensitive operations can introduce confirmation and policy-driven reauthentication.

A missing action helps the UI render correctly; it is not an authorization boundary.

## Management policy

Proposed default policy:

| Operation | Owner | Admin | Member |
| --- | --- | --- | --- |
| List members and roles | Yes | Yes | Yes |
| Rename organization | Yes | Yes | No |
| List, send, resend, revoke member invitations | Yes | Yes | No |
| Remove an ordinary member | Yes | Yes | No |
| Appoint/demote an admin | Yes | No | No |
| Remove an admin | Yes | No | No |
| Transfer ownership | Yes | No | No |
| Leave organization | After transfer | Yes | Yes |

Invitation acceptance grants Member only. Elevation is a separate owner-authorized operation. Admins cannot affect other admins or the owner, including through crafted target IDs or alternative actions.

Do not assign special permissions to existing arbitrary roles merely because their slugs happen to be `owner`, `admin`, or `member`. Recommended direction: explicit protected management-role IDs configured per tenant, mapped to the existing role definitions/bindings. A managed organization has exactly one management role per member; unrelated application roles may coexist.

Store a canonical owner reference for managed organizations, enforce that the owner is a member of the same organization, and update its protected role binding in the same transaction. Define one authoritative source for ownership and keep its role-check projection consistent. Existing organizations remain unmanaged until explicitly adopted with a selected owner.

Admin API adapters must respect managed-organization invariants too. Removing the owner membership, deleting a protected role, or deleting an owner account cannot silently leave an invalid organization. Account deletion requires ownership transfer or an explicit organization-deletion policy; this must be resolved before enabling the feature.

## Invitation lifecycle

An invitation is durable data, not a long-lived Flow API instance.

Suggested fields: ID, tenant ID, organization ID, normalized recipient email, inviter user ID, token digest, expiry, acceptance/revocation timestamps, accepting user ID, and delivery metadata. Separate lifecycle state (`pending`, `accepted`, `revoked`, `expired`) from delivery state (queued/sent/failed). The UI displays joined members separately from outstanding invitations and can show historical accepted invitations where authorized.

Normalize addresses according to Hanko's existing email rules; do not introduce provider-specific alias stripping. Prevent duplicate active invitations for the same tenant/organization/email.

Acceptance proceeds as follows:

1. Exchange an opaque, high-entropy invitation token into a short-lived flow context. Store only its digest durably. Avoid logging tokens or leaking them through referrers.
2. If no authenticated user exists, require login/registration using the existing flows. Keep invitation continuation server-side and use only allowlisted return destinations.
3. Bind the resumed flow to the authenticated user and verify that an account email is verified and matches the invited address. A different signed-in account must not consume the invitation.
4. Present the organization and an explicit acceptance action.
5. Recheck invitation validity and policy, then atomically add membership, grant Member, and mark the invitation accepted.

The bearer invitation token alone must not grant membership. Do not publish member lists or account-existence information in unauthenticated invitation states. Merely opening an email link must not accept an invitation; mail scanners can follow links.

Expired flows can be restarted from an unexpired invitation. Expired/revoked invitations cannot be revived by replaying an earlier flow action. Resending rotates the token and invalidates old links. Repeated acceptance by the same user returns the already-joined outcome; it never duplicates or downgrades an existing membership.

Recheck the inviter's current authority on acceptance: under this proposal, an inviter who no longer has invitation permission cannot leave an invitation capable of granting access. The UI should expose that invalidation and allow an authorized administrator to issue a fresh invitation.

Reuse Hanko's mail/template infrastructure where possible, with localized invitation templates. Commit invitations and a delivery outbox atomically; send only after commit, retry delivery idempotently, and limit send/resend rates by tenant, actor, organization, and recipient.

## Ownership transfer and concurrency

Suggested flow:

`select_member → confirm_transfer → optional_reauthentication → transfer_complete`

Bind any reauthentication evidence to the actor, organization, target, and transfer operation with a short validity period. The target must still be a member when the action commits.

Within one database transaction, lock or version-check the organization and affected memberships, confirm that the caller is still the current owner, change the canonical owner, grant the new owner role, and demote the old owner to Admin. Concurrent transfer/removal/role-change requests must not produce zero or multiple owners.

Flow-level locking alone is insufficient: another flow or an Admin API call may modify the same organization. Define a common database locking order and conflict behavior across all mutation paths. Return a refreshable conflict state for stale UI actions; do not continue with cached permissions.

## Personal organization provisioning

A tenant option enables once-only creation when a new account is created, including supported passwordless, password, and social/enterprise registration paths. An invited new user still receives their own organization before joining the inviting organization.

Use a durable provisioning identity unique per tenant and original user, independent of current ownership. Transferring a personal organization must not trigger a replacement on the next login. Provisioning and initial membership/ownership must be atomic or recover through an explicitly idempotent process.

For existing users, provide an explicit backfill operation; do not make every login implicitly create organizations. Decide separately whether administrative user creation/import should provision by default or require an explicit option.

To preserve the current required, tenant-unique `name` contract, initially keep a generated unique internal name and add an optional `display_name` for self-service organizations. The UI uses a localized fallback such as "Personal organization" when it is absent. Duplicate display names are allowed; URLs and authorization always use IDs. This avoids filling the current unique name column with the same default for every user.

## Shared implementation boundary

Flow handlers and public read handlers call a shared organization service. Admin handlers reuse the same mutation/invariant layer with an explicit privileged actor context; they must not call the public flow over HTTP or bypass constraints through direct writes.

For every action:

- Resolve tenant from trusted routing context and actor from the validated session. Resolve public user IDs to internal IDs as the current role-check handler does.
- Bind flow state to tenant, actor, and organization; reject cross-context replay.
- Revalidate session, current membership, and permissions against persisted data at execution, not only flow initialization or stale JWT role claims.
- Validate all targets in the same tenant and organization.
- Preserve Flow API CSRF/origin protections and reject arbitrary client-selected return URLs.
- Make retries safe and record auditable actor/target/outcome events without tokens or unnecessary personal data.

All these checks apply to reads and mutations as appropriate. CORS and omitted UI buttons are not authorization controls.

## SDK and Elements

Extend SDK flow types, action/state payload types, and error handling alongside the backend schema. Keep the existing profile flow focused on the individual account.

A future `<hanko-organization>` can render management states and paginated member/invitation lists. Invitation acceptance should also work as a standalone view for email deep links. Expose events for organization updates, membership changes, and accepted invitations; applications own routing and their selected-organization UI.

Selecting an organization does not automatically authorize access to an application's own data. Downstream applications still need server-side authorization for their protected resources.

## Rollout and validation

Suggested implementation sequence:

1. Finalize role/owner representation, configuration, DTOs, and compatibility rules.
2. Add migrations, shared authorization/invariants, and personal-org provisioning.
3. Add authorized read endpoints and management flows.
4. Add durable invitations, delivery, and acceptance continuation.
5. Add SDK support, Elements, translations, examples, and API documentation.

Before enabling the feature, test:

- All matrix entries, anonymous callers, arbitrary action URLs, and cross-tenant/cross-organization IDs.
- Permission removal between flow initialization and action execution, including session revocation.
- Replayed/expired flow state, CSRF failures, and invitation continuation after authentication.
- Wrong/unverified email, changed recipient email, revoked/expired/rotated tokens, and already-joined users.
- Concurrent invitation acceptance, ownership transfer, membership removal, and role deletion.
- Duplicate delivery retries and persisted invitation state after mail failures.
- Exactly-once personal provisioning across retries and registration paths, and no reprovisioning after transfer.
- Existing generic roles and unmanaged organizations, owner account deletion, and Admin API invariants.
- Pagination and privacy of list responses; custom frontend and Element paths.

This documentation-only PR runs no backend or browser tests and makes no runtime changes. These cases are acceptance criteria for subsequent implementation.

## Decisions requested

- Approve the hybrid read-endpoint/Flow API surface and standalone invitation flow?
- Confirm explicit protected management roles plus canonical ownership, rather than inferring privileges from arbitrary role slugs?
- Confirm generated internal names plus optional display names for compatibility?
- Confirm Member-only invitations, inviter-authority rechecks, and previous-owner demotion to Admin?
- Choose account-deletion handling, provisioning behavior for imports/admin-created users, retention of invitation history, and reauthentication defaults before implementation.
