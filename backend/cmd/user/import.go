package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

var inputFile string

func NewImportCommand() *cobra.Command {
	var (
		configFile string
	)

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import users into database from a Json file",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			//Load cfg
			cfg, err := config.Load(&configFile)
			if err != nil {
				log.Fatal(err)
			}

			//Load File
			users, err := loadFile()
			if err != nil {
				log.Println(err)
				os.Exit(1)
			}
			//Validate Input
			err = validateEntries(users)
			if err != nil {
				log.Println(err)
				os.Exit(1)
			}
			//Import Users
			persister, err := persistence.New(cfg.Database)
			if err != nil {
				log.Println(err)
				os.Exit(1)
			}
			err = addToDatabase(users, persister)
			if err != nil {
				log.Println(err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVar(&configFile, "config", config.DefaultConfigFilePath, "config file")
	cmd.Flags().StringVarP(&inputFile, "inputFile", "i", "", "The json file where the users should be imported from.")
	err := cmd.MarkFlagRequired("inputFile")
	if err != nil {
		log.Println(err)
	}
	return cmd
}

func loadFile() ([]ImportEntry, error) {
	jsonFile, err := os.Open(inputFile)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)

	var users []ImportEntry
	err = json.Unmarshal(byteValue, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func addToDatabase(entries []ImportEntry, persister persistence.Persister) error {
	tx := persister.GetConnection()
	err := tx.Transaction(func(tx *pop.Connection) error {
		for i, v := range entries {
			now := time.Now().UTC()
			// pre genereate a v4 uuid
			userId, _ := uuid.NewV4()

			// if there is an userId set try to parse into uuid
			if v.UserID != "" {
				err := userId.Parse(v.UserID)
				if err != nil {
					return errors.New(fmt.Sprintf("Error Adding entry nr. %v. Error Parsing as uuid: %v", i, v.UserID))
				}
			}
			createdAt := now
			updatedAt := createdAt
			if v.CreatedAt != nil {
				createdAt = *v.CreatedAt
			}
			if v.UpdatedAt != nil {
				updatedAt = *v.UpdatedAt
			}

			u := models.User{
				ID:        userId,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			}

			err := tx.Create(&u)
			if err != nil {
				return fmt.Errorf("Failed to create user with id: %v : %w", u.ID.String(), err)
			}

			for _, e := range v.Emails {
				emailId, _ := uuid.NewV4()

				mail := models.Email{
					ID:        emailId,
					UserID:    &userId,
					Address:   strings.ToLower(e.Address),
					Verified:  e.IsVerified,
					CreatedAt: now,
					UpdatedAt: now,
				}
				err := tx.Create(&mail)
				if err != nil {
					return fmt.Errorf("Failed to create email %v for user %v : %w", e.Address, userId.String(), err)
				}

				if e.IsPrimary {
					primary := &models.PrimaryEmail{
						UserID:  userId,
						EmailID: emailId,
					}
					err = tx.Create(primary)
					if err != nil {
						return fmt.Errorf("Failed to set email %v as  primary for user %v : %w", e.Address, userId.String(), err)
					}
				}
			}
		}
		return nil
	})

	return err
}
