# Keycloak realm import

Every `*-realm.json` file in this directory is loaded automatically on Keycloak
startup (`--import-realm` in `quickstart.keycloak.yaml`), so realm/client/mapper/user
setup done once in the admin console is available to every dev without repeating the
manual steps.

## Updating a realm export

Keycloak can't export a realm from its default embedded storage without stopping the
server, but this setup backs Keycloak with a real Postgres database
(`keycloak-db-init` in `quickstart.keycloak.yaml`), so the export can run as a
separate, short-lived container against that same database while the main
`keycloak` container keeps running - no downtime, no need to stop anything first.

1. Make your changes in the Keycloak admin console (`http://localhost:9000`) as usual
   - new realms, clients, mappers, users, attributes, etc.
2. Export, from `deploy/docker-compose/`:

   ```
   docker compose -f quickstart.yaml -f quickstart.keycloak.yaml run --rm keycloak export --dir=/opt/keycloak/data/import --realm=uni-a --users=realm_file
   ```

   - `--dir=/opt/keycloak/data/import` writes straight into this directory (already
     bind-mounted), so the updated file lands here directly.
   - `--realm=uni-a` - change this if you're exporting a different/additional realm.
     If you've changed multiple realms, omit `--realm` entirely - Keycloak then
     exports every realm, each into its own `<realm-name>-realm.json` file in
     `--dir`, which is exactly what `--import-realm` expects on the next startup.
   - `--users=realm_file` bundles users (with their hashed credentials and attributes)
     into the same file, rather than skipping them or splitting them into separate files.
   - Use `=` for every flag value, not a space. A pasted `--flag value` can get split
     across a line break by some terminals, silently dropping the value (seen in
     practice with `--users`) - `--flag=value` can't be split that way.
   - If this fails with a port/container-name conflict, your running stack is probably
     under a different Compose project name than the default. Check with
     `docker compose ls`, then add `-p <that name>` right after `docker compose` in the
     command above.
3. Commit the updated `*-realm.json` file.