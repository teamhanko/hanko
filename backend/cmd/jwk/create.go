package jwk

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gofrs/uuid"
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/crypto/jwk"
	"github.com/teamhanko/hanko/backend/v3/crypto/jwk/local_db"
	"github.com/teamhanko/hanko/backend/v3/persistence"
)

func NewCreateCommand() *cobra.Command {
	var (
		configFile string
		tenantID   string
		store      bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "generate a JSON Web Key and print it, or persist it for a tenant with --store",
		Long: `Generates a JWK and prints it to the console.

With --store, the JWK is instead encrypted with the tenant's configured secret(s) and persisted to the
database, the same way a tenant's initial JWK is created via the API.

Local, single-tenant dev server (e.g. via docker-compose): no flags needed, defaults match the local dev config:

  hanko jwk create --store

Against a running multi-tenant deployment (e.g. in a Kubernetes cluster): run it inside the existing server pod
instead of setting up a local DB connection, so it reuses the pod's already-mounted config and DB connectivity:

  kubectl exec -n <namespace> deploy/hanko -- /hanko jwk create --store --config /etc/config/config.yaml --tenant_id <tenant_id>
`,
		Run: func(cmd *cobra.Command, args []string) {
			if !store {
				generator := local_db.RSAKeyGenerator{}
				key, err := generator.Generate("key1")
				if err != nil {
					log.Panicln(err)
				}
				j, err := json.Marshal(key)
				if err != nil {
					log.Panicln(err)
				}
				fmt.Println(string(j))
				return
			}

			cfg, err := config.Load(&configFile)
			if err != nil {
				log.Fatal(err)
			}

			dbConnection, err := persistence.NewConnection(cfg.Database)
			if err != nil {
				log.Fatal(err)
			}
			persister := persistence.New(dbConnection)

			var tID uuid.UUID
			if !cfg.ApplicationConfig.MultiTenancy.Enabled {
				tID = uuid.FromStringOrNil(config.DefaultTenantID)
			} else {
				if tenantID == "" {
					log.Fatal("tenant_id must be present if multitenancy is enabled")
				}
				tID, err = uuid.FromString(tenantID)
				if err != nil {
					log.Fatalf("invalid tenant_id: %s", err)
				}

				tenantModel, err := persister.GetTenantPersister().Get(tID)
				if err != nil {
					log.Fatalf("failed to load tenant: %s", err)
				}
				if tenantModel == nil {
					log.Fatalf("tenant with id '%s' not found", tID.String())
				}

				tenantConfig, err := config.ParseMultitenancyTenantConfig(tenantModel.Config)
				if err != nil {
					log.Fatalf("failed to load tenant config: %s", err)
				}
				cfg.TenantConfig = *tenantConfig
			}

			if err = cfg.PostProcess(); err != nil {
				log.Fatalf("failed to post process config: %s", err)
			}

			jwkManager, err := jwk.NewManager(*cfg, persister)
			if err != nil {
				log.Fatalf("failed to create jwk manager: %s", err)
			}

			key, err := jwkManager.GenerateKey(tID)
			if err != nil {
				log.Fatalf("failed to generate and persist jwk: %s", err)
			}

			fmt.Printf("Generated and persisted JWK '%s' for tenant '%s'\n", key.KeyID(), tID.String())
		},
	}

	cmd.Flags().BoolVar(&store, "store", false, "encrypt and persist the generated JWK for a tenant instead of printing it")
	cmd.Flags().StringVar(&configFile, "config", "", "config file (only used with --store)")
	cmd.Flags().StringVar(&tenantID, "tenant_id", "", "tenant ID, required if multitenancy is enabled (only used with --store)")

	return cmd
}
