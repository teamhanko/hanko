package certificate

import (
	"errors"
	"fmt"
	"log"

	"github.com/gofrs/uuid"
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/persistence"
)

func NewCreateCommand() *cobra.Command {
	var (
		configFile string
		tenantID   string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "create a SAML certificate for a tenant that doesn't have one yet",
		Long:  ``,
		Args: func(cmd *cobra.Command, args []string) error {
			if tenantID == "" {
				return errors.New("tenant_id is required")
			}
			if _, err := uuid.FromString(tenantID); err != nil {
				return errors.New("tenant_id is not a uuid")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.Load(&configFile)
			if err != nil {
				log.Fatal(err)
			}

			tID, err := uuid.FromString(tenantID)
			if err != nil {
				log.Fatalf("invalid tenant_id: %s", err)
			}

			dbConnection, err := persistence.NewConnection(cfg.Database)
			if err != nil {
				log.Fatal(err)
			}
			dbPersister := persistence.New(dbConnection)

			result, err := createCertificateForTenant(dbPersister, tID, false)
			if err != nil {
				log.Fatalf("failed to create certificate: %s", err)
			}
			fmt.Println(result.Message)
		},
	}

	cmd.Flags().StringVar(&configFile, "config", config.DefaultConfigFilePath, "config file")
	cmd.Flags().StringVar(&tenantID, "tenant_id", "", "tenant ID to create a SAML certificate for")

	return cmd
}
