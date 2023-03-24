package user

import (
	"encoding/json"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gofrs/uuid"
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/config"
	"log"
	"os"
	"time"
)

var outputFile string
var count int

func NewGenerateCommand(config *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate mock users and write them to a file.",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			generate()
		},
	}

	cmd.Flags().StringVarP(&outputFile, "outputFile", "o", "", "The path of the output file.")
	err := cmd.MarkFlagRequired("outputFile")
	if err != nil {
		log.Println(err)
	}
	cmd.Flags().IntVarP(&count, "count", "c", 10, "Gives the number of users that should be generated.")
	return cmd
}

func generate() error {
	var entries []ImportEntry
	for i := 0; i < count; i++ {
		now := time.Now().UTC()
		id, _ := uuid.NewV4()
		emails := []ImportEmail{
			{
				Address:    gofakeit.Email(),
				IsPrimary:  true,
				IsVerified: true,
			},
		}
		entry := ImportEntry{
			UserID:    id.String(),
			Emails:    emails,
			CreatedAt: &now,
			UpdatedAt: &now,
		}
		entries = append(entries, entry)
	}
	bytes, err := json.Marshal(entries)
	if err != nil {
		return err
	}
	err = os.WriteFile(outputFile, bytes, 0644)
	if err != nil {
		return err
	}

	return nil
}
