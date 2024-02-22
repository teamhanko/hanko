package user

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gofrs/uuid"
	"github.com/spf13/cobra"
)

var outputFile string
var count int

func NewGenerateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate mock users and write them to a file.",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			err := generate()
			if err != nil {
				log.Println(err)
			}
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
	var entries []ImportOrExportEntry
	for i := 0; i < count; i++ {
		now := time.Now().UTC()
		id, _ := uuid.NewV4()
		emails := []ImportOrExportEmail{
			{
				Address:    gofakeit.Email(),
				IsPrimary:  true,
				IsVerified: true,
			},
		}
		entry := ImportOrExportEntry{
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
	err = os.WriteFile(outputFile, bytes, 0600)
	if err != nil {
		return err
	}

	return nil
}
