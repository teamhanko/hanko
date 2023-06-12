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

func NewImportCommand() *cobra.Command {
	var (
		configFile string
		inputFile  string
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
			// we explicitly want user input here, hence  #nosec G304
			jsonFile, err := os.Open(inputFile)
			if err != nil {
				log.Fatal(err)
			}
			defer func() {
				if err := jsonFile.Close(); err != nil {
					log.Printf("Error closing file: %s\n", err)
				}
			}()
			users, err := loadAndValidate(jsonFile)
			if err != nil {
				log.Fatal(err)
			}
			//Import Users
			persister, err := persistence.New(cfg.Database)
			if err != nil {
				log.Fatal(err)
			}
			err = addToDatabase(users, persister)
			if err != nil {
				log.Fatal(err)
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

// loadAndValidate reads json from an io.Reader so we read every entry separate and validate it. We go through the whole
// array to print out every validation error in the input data.
func loadAndValidate(input io.Reader) ([]ImportEntry, error) {
	dec := json.NewDecoder(input)

	// read the open bracket
	_, err := dec.Token()
	if err != nil {
		return nil, err
	}

	users := []ImportEntry{}

	numErrors := 0
	index := 0
	// while the array contains values
	for dec.More() {
		index = index + 1
		var userEntry ImportEntry
		// decode one ImportEntry
		err := dec.Decode(&userEntry)
		if err != nil {
			errorMsg := fmt.Sprintf("Error at entry %v : %v", index, err.Error())
			log.Println(errorMsg)
			return nil, err
		}

		if err := userEntry.validate(); err != nil {
			errorMsg := fmt.Sprintf("Error at entry %v : %v", index, err.Error())
			log.Println(errorMsg)
			log.Print(userEntry)
			numErrors++
			continue
		}
		users = append(users, userEntry)
	}

	// read closing bracket
	_, err = dec.Token()
	if err != nil {
		return nil, err
	}
	if numErrors > 0 {
		errMsg := fmt.Sprintf("Found %v errors.", numErrors)
		log.Printf(errMsg)
		return nil, errors.New(errMsg)
	}

	return users, nil
}

// commits the list of ImportEntries to the database. Wrapped in a transaction so if something fails no new users are added.
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
