package user

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gofrs/uuid"
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/persistence"
)

// TODO: should include a tenantID as parameter
func NewExportCommand() *cobra.Command {
	var (
		configFile string
		outputFile string
		tenantID   string
	)

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export users from database into a Json file",
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

			var tID uuid.UUID
			if tenantID != "" {
				tID, err = uuid.FromString(tenantID)
				if err != nil {
					log.Fatalf("invalid tenant_id: %s", err)
				}
			}
			persister := persistence.New(dbConnection)

			err = export(persister, outputFile, &tID)
			if err != nil {
				log.Fatal(err)
			}
			log.Println(fmt.Sprintf("Successfully exported users to %s", outputFile))
		},
	}

	cmd.Flags().StringVar(&configFile, "config", config.DefaultConfigFilePath, "config file")
	cmd.Flags().StringVarP(&outputFile, "outputFile", "o", "", "The path of the output file.")
	cmd.Flags().StringVar(&tenantID, "tenant_id", "", "tenant ID (optional)")
	err := cmd.MarkFlagRequired("outputFile")
	if err != nil {
		log.Println(err)
	}
	return cmd
}

func export(persister persistence.Persister, outFile string, tenantID *uuid.UUID) error {
	var entries []ImportOrExportEntry
	users, err := persister.GetUserPersister().All(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get list of users: %w", err)
	}
	for _, user := range users {
		var emails []ImportOrExportEmail
		for _, email := range user.Emails {
			emails = append(emails, ImportOrExportEmail{
				Address:    email.Address,
				IsPrimary:  email.IsPrimary(),
				IsVerified: email.Verified,
			})
		}
		entry := ImportOrExportEntry{
			UserID:    user.ID.String(),
			Emails:    emails,
			CreatedAt: &user.CreatedAt,
			UpdatedAt: &user.UpdatedAt,
		}
		entries = append(entries, entry)
	}
	bytes, err := json.Marshal(entries)
	if err != nil {
		return err
	}
	err = os.WriteFile(outFile, bytes, 0600)
	if err != nil {
		return err
	}
	return nil
}
