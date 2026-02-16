# Multi-Tenant Docker Deployment & Testing Guide

This guide covers how to run the updated Hanko with multi-tenant support in Docker and test the feature.

## Table of Contents
- [Quick Reference Commands](#quick-reference-commands)
- [Step-by-Step Deployment](#step-by-step-deployment)
- [Configuration](#configuration)
- [Testing Multi-Tenant Feature](#testing-multi-tenant-feature)
- [Troubleshooting](#troubleshooting)

---

## Quick Reference Commands

```bash
# In-place upgrade (preserves data, runs migrations)
docker compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart" down
docker compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart" up --build

# Fresh start (clears database)
docker compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart" down -v
docker compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart" up --build

# Rebuild Hanko only (config changes, no DB changes)
docker compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart" up -d --build --no-deps hanko

# Restart Hanko (config reload only)
docker compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart" restart hanko

# View logs
docker compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart" logs -f hanko

# Run migrations manually
docker compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart" run --rm hanko-migrate
```

---

## Step-by-Step Deployment

### Deployment Options

The multi-tenant migrations are **additive and backward-compatible**. You have two options:

| Approach | When to Use | Data Preserved |
|----------|-------------|----------------|
| **Option A: In-Place Migration** | Production, existing users | ✅ Yes |
| **Option B: Fresh Start** | Development, testing | ❌ No |

---

### Option A: In-Place Migration (Recommended for Production)

This approach preserves all existing users and data. The migrations will:
- Create the new `tenants` table
- Add `tenant_id` column to existing tables (nullable, defaults to `NULL`)
- Existing users become "global users" (`tenant_id = NULL`)
- Global users are automatically adopted into tenants when they log in with `X-Tenant-ID` header

#### Step A1: Stop Containers (Keep Data)

```bash
# Stop containers WITHOUT removing volumes (preserves DB)
docker compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart" down
```

#### Step A2: Update Configuration

Edit `deploy/docker-compose/config.yaml` and add the multi-tenant configuration:

```yaml
multi_tenant:
  enabled: true
  tenant_header: "X-Tenant-ID"
  allow_global_users: true    # IMPORTANT: Keep true to support existing users
  auto_provision: true        # Auto-create tenants on first request
```

> 💡 **Note**: Setting `allow_global_users: true` is essential for backward compatibility with existing users.

#### Step A3: Build and Run with Migration

```bash
# Rebuild and run - migrations will apply automatically
docker compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart" up --build
```

The migrations will:
1. Add `tenants` table
2. Add `tenant_id` column to users, emails, usernames, etc.
3. Create partial unique indexes for tenant-scoped uniqueness
4. **Existing data remains intact** with `tenant_id = NULL`

#### Step A4: Verify Migration

```bash
# Check migration logs
docker compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart" logs hanko-migrate
```

Verify existing users still work:
```bash
# Existing users can still log in without X-Tenant-ID header
curl -X POST http://localhost:8000/login \
  -H "Content-Type: application/json" \
  -d '{}'
```

---

### Option B: Fresh Start (Development/Testing)

Use this approach when you want to start with a clean database.

#### Step B1: Stop and Clear Everything

```bash
# Stop and remove containers AND volumes (clears DB)
docker compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart" down -v
```

> ⚠️ **Warning**: The `-v` flag removes volumes, which **permanently deletes** all database data.

#### Step B2: Update Configuration

Edit `deploy/docker-compose/config.yaml` and add the multi-tenant configuration:

```yaml
multi_tenant:
  enabled: true
  tenant_header: "X-Tenant-ID"
  allow_global_users: true    # Allow users without tenant
  auto_provision: true        # Auto-create tenants on first request
```

#### Step B3: Build and Run

```bash
# Full rebuild with fresh database
docker compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart" up --build
```

---

### What Happens During Migration

Both approaches run the same migrations:

| Migration | Description |
|-----------|-------------|
| `create_tenants` | Creates `tenants` table for storing tenant info |
| `add_tenant_id` | Adds nullable `tenant_id` column to users, emails, usernames, etc. |
| `change_unique_constraints` | Creates partial unique indexes for tenant-scoped uniqueness |

**Existing data behavior:**
- All existing records retain their data with `tenant_id = NULL`
- These become "global users" that work without tenant context
- No data is lost or modified beyond adding the new column

### Global User Auto-Adoption Flow

When a global user (existing user with `tenant_id = NULL`) logs in with an `X-Tenant-ID` header:

```
1. User submits login with email + X-Tenant-ID header
2. System looks for email in specified tenant → Not found
3. System falls back to global users (tenant_id = NULL) → Found!
4. System "adopts" user into tenant:
   - Updates user.tenant_id = <tenant_id>
   - Updates emails.tenant_id = <tenant_id>
   - Updates all related records (credentials, sessions, etc.)
5. Login completes successfully
6. User is now part of the tenant
```

This ensures seamless migration without requiring users to re-register.

### Verify Migration Success

Check the logs to ensure migrations ran successfully:

```bash
docker compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart" logs hanko-migrate
```

You should see output indicating the migrations were applied:
```
[migrate] Applied migration: 20260120000001_create_tenants
[migrate] Applied migration: 20260120000002_add_tenant_id
[migrate] Applied migration: 20260120000003_change_unique_constraints
```

---

## Configuration

### Multi-Tenant Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `enabled` | `false` | Enable multi-tenant mode |
| `tenant_header` | `X-Tenant-ID` | HTTP header name for tenant ID |
| `allow_global_users` | `true` | Allow users without tenant (backward compatible) |
| `auto_provision` | `true` | Auto-create tenant if ID doesn't exist |

### Example Configurations

**Development (Multi-tenant enabled with auto-provisioning):**
```yaml
multi_tenant:
  enabled: true
  tenant_header: "X-Tenant-ID"
  allow_global_users: true
  auto_provision: true
```

**Production (Strict mode - tenant required, no auto-provisioning):**
```yaml
multi_tenant:
  enabled: true
  tenant_header: "X-Tenant-ID"
  allow_global_users: false
  auto_provision: false
```

**Disabled (Single-tenant mode - default):**
```yaml
multi_tenant:
  enabled: false
```

---

## Testing Multi-Tenant Feature

### Prerequisites
- Hanko running on `http://localhost:8000` (public API)
- Hanko Admin API running on `http://localhost:8001`
- MailSlurper running on `http://localhost:8080` (email UI)

### Test 1: Register User Without Tenant (Backward Compatibility)

```bash
# Initialize registration flow (no tenant header)
curl -X POST http://localhost:8000/registration \
  -H "Content-Type: application/json" \
  -d '{}'
```

This should work when `allow_global_users: true`.

### Test 2: Register User With Tenant (Auto-Provisioning)

```bash
# Generate a random tenant UUID
TENANT_ID=$(uuidgen)
echo "Using Tenant ID: $TENANT_ID"

# Initialize registration flow with tenant
curl -X POST http://localhost:8000/registration \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{}'
```

The tenant will be auto-created if `auto_provision: true`.

### Test 3: Same Email in Different Tenants

This is the key feature - the same email can exist in multiple tenants:

```bash
# Tenant A
TENANT_A="11111111-1111-1111-1111-111111111111"

# Tenant B
TENANT_B="22222222-2222-2222-2222-222222222222"

# Register user@example.com in Tenant A
curl -X POST http://localhost:8000/registration \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_A" \
  -d '{}'
# Complete the flow with email: user@example.com

# Register SAME email in Tenant B (should succeed!)
curl -X POST http://localhost:8000/registration \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_B" \
  -d '{}'
# Complete the flow with email: user@example.com
```

### Test 4: Verify Tenant Isolation via Admin API

```bash
# List all tenants
curl -X GET http://localhost:8001/tenants

# Get specific tenant
curl -X GET http://localhost:8001/tenants/$TENANT_A

# List users in a tenant
curl -X GET http://localhost:8001/tenants/$TENANT_A/users
```

### Test 5: Login with Tenant Context

```bash
# Login as user in Tenant A
curl -X POST http://localhost:8000/login \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_A" \
  -d '{}'

# Same user, different tenant should be different account
curl -X POST http://localhost:8000/login \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_B" \
  -d '{}'
```

### Test 5b: Global User Auto-Adoption (In-Place Migration Scenario)

If you upgraded from a single-tenant setup, existing users can be adopted into tenants:

```bash
# 1. First, verify user exists without tenant (global user)
docker exec -it hanko-quickstart-postgresd-1 psql -U hanko -d hanko -c \
  "SELECT id, tenant_id FROM users WHERE id IN (SELECT user_id FROM emails WHERE address = 'existing@example.com');"
# Should show tenant_id = NULL

# 2. Login with X-Tenant-ID header - triggers auto-adoption
NEW_TENANT="33333333-3333-3333-3333-333333333333"
curl -X POST http://localhost:8000/login \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $NEW_TENANT" \
  -d '{}'
# Complete the login flow with existing@example.com

# 3. Verify user was adopted into tenant
docker exec -it hanko-quickstart-postgresd-1 psql -U hanko -d hanko -c \
  "SELECT id, tenant_id FROM users WHERE id IN (SELECT user_id FROM emails WHERE address = 'existing@example.com');"
# Should now show tenant_id = 33333333-3333-3333-3333-333333333333
```

### Test 6: Verify JWT Contains tenant_id

After successful login, decode the JWT token. It should contain:
```json
{
  "sub": "user-uuid",
  "tenant_id": "tenant-uuid",
  "email": "user@example.com",
  ...
}
```

### Test 7: Admin API - Create Tenant Manually

```bash
curl -X POST http://localhost:8001/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "name": "ClinicOS Tenant",
    "slug": "clinicos-main"
  }'
```

### Test 8: Admin API - Update Tenant

```bash
curl -X PUT http://localhost:8001/tenants/$TENANT_ID \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Tenant Name",
    "enabled": true
  }'
```

### Test 9: Disabled Tenant Returns 403

```bash
# Disable tenant
curl -X PUT http://localhost:8001/tenants/$TENANT_ID \
  -H "Content-Type: application/json" \
  -d '{"enabled": false}'

# Try to access with disabled tenant
curl -X POST http://localhost:8000/login \
  -H "X-Tenant-ID: $TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{}'
# Should return 403 Forbidden
```

---

## Database Verification

Connect to PostgreSQL to verify the schema:

```bash
# Connect to PostgreSQL container
docker exec -it hanko-quickstart-postgresd-1 psql -U hanko -d hanko

# Check tenants table
\d tenants
SELECT * FROM tenants;

# Check tenant_id column on users
\d users
SELECT id, tenant_id, created_at FROM users;

# Check tenant_id column on emails
\d emails
SELECT id, address, tenant_id FROM emails;

# Check unique indexes
\di emails_address_tenant_idx
\di emails_address_global_idx
```

---

## Troubleshooting

### Migration Errors

If you see migration errors:

```bash
# Check migration status
docker compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart" logs hanko-migrate

# Reset everything and retry
docker compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart" down -v
docker compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart" up --build
```

### "Tenant not found" Error

If you get a 404 error for tenant:
- Check that `auto_provision: true` is set in config
- Verify the tenant ID is a valid UUID format
- Check that multi-tenant mode is enabled: `multi_tenant.enabled: true`

### "X-Tenant-ID header required" Error

This occurs when:
- `allow_global_users: false` is set
- No `X-Tenant-ID` header is provided

Solution: Either provide the header or set `allow_global_users: true`.

### Build Errors

```bash
# Clean rebuild
docker compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart" build --no-cache hanko
docker compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart" up
```

### View Real-time Logs

```bash
# All services
docker compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart" logs -f

# Hanko only
docker compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart" logs -f hanko

# Migrations only
docker compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart" logs hanko-migrate
```

---

## Endpoints Summary

### Public API (port 8000)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/registration` | POST | Start registration flow |
| `/login` | POST | Start login flow |
| `/profile` | POST | Profile management flow |

All endpoints accept `X-Tenant-ID` header when multi-tenant is enabled.

### Admin API (port 8001)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/tenants` | GET | List all tenants |
| `/tenants` | POST | Create new tenant |
| `/tenants/:id` | GET | Get tenant by ID |
| `/tenants/:id` | PUT | Update tenant |
| `/tenants/:id` | DELETE | Delete tenant |
| `/tenants/:id/users` | GET | List users in tenant |

---
