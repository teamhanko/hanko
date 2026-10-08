package certificate

import (
	"fmt"
	"log"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/persistence"
)

func NewCreateAllCommand() *cobra.Command {
	var (
		configFile string
		dryRun     bool
	)

	cmd := &cobra.Command{
		Use:   "create-all",
		Short: "create a SAML certificate for every tenant that doesn't have one yet",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.Load(&configFile)
			if err != nil {
				log.Fatal(err)
			}

			dbConnection, err := persistence.NewConnection(cfg.Database)
			if err != nil {
				log.Fatal(err)
			}
			dbPersister := persistence.New(dbConnection)

			affected, err := findTenantsMissingCertificate(dbConnection)
			if err != nil {
				log.Fatalf("failed to find affected tenants: %s", err)
			}

			if dryRun {
				fmt.Println("DRY RUN - no data will be written")
			}
			fmt.Printf("%d tenant(s) missing a SAML certificate\n", len(affected))

			created, skipped, failed := 0, 0, 0
			for _, tenantID := range affected {
				result, err := createCertificateForTenant(dbPersister, tenantID, dryRun)
				if err != nil {
					fmt.Printf("  tenant %s: FAILED - %s\n", tenantID, err)
					failed++
					continue
				}

				fmt.Printf("  %s\n", result.Message)
				switch result.Outcome {
				case outcomeCreated, outcomeWouldCreate:
					created++
				default:
					skipped++
				}
			}

			fmt.Printf("\ndone: %d created, %d skipped, %d failed\n", created, skipped, failed)
		},
	}

	cmd.Flags().StringVar(&configFile, "config", config.DefaultConfigFilePath, "config file")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "report what would be created without writing anything")

	return cmd
}

type tenantRow struct {
	ID uuid.UUID `db:"id"`
}

func findTenantsMissingCertificate(db *pop.Connection) ([]uuid.UUID, error) {
	var rows []tenantRow
	err := db.RawQuery(`
		SELECT t.id FROM tenants t
		WHERE NOT EXISTS (SELECT 1 FROM saml_certificates c WHERE c.tenant_id = t.id)
	`).All(&rows)
	if err != nil {
		return nil, fmt.Errorf("failed to query affected tenants: %w", err)
	}

	ids := make([]uuid.UUID, len(rows))
	for i, r := range rows {
		ids[i] = r.ID
	}
	return ids, nil
}
