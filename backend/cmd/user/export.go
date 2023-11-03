package user

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/persistence"
)

func NewExportCommand() *cobra.Command {
	var (
		configFile string
		outputFile string
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
			persister, err := persistence.New(cfg.Database)
			if err != nil {
				log.Fatal(err)
			}
			err = export(persister, outputFile)
			if err != nil {
				log.Fatal(err)
			}
			log.Println(fmt.Sprintf("Successfully exported users to %s", outputFile))
		},
	}

	cmd.Flags().StringVar(&configFile, "config", config.DefaultConfigFilePath, "config file")
	cmd.Flags().StringVarP(&outputFile, "outputFile", "o", "", "The path of the output file.")
	err := cmd.MarkFlagRequired("outputFile")
	if err != nil {
		log.Println(err)
	}
	return cmd
}

func export(persister persistence.Persister, outFile string) error {
	var entries []ImportOrExportEntry
	users, err := persister.GetUserPersister().All()
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
