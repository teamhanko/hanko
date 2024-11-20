package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/persistence"
)

func NewImportCommand() *cobra.Command {
	var (
		configFile string
		inputFile  string
		inputUrl   string
	)

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import users into database from a Json file",
		Long:  ``,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			fileFlagSet := cmd.Flags().Changed("inputFile")
			urlFlagSet := cmd.Flags().Changed("inputUrl")
			if !fileFlagSet && !urlFlagSet {
				return errors.New("either flag \"inputFile\" or \"inputUrl\" must be set")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			//Load cfg
			cfg, err := config.Load(&configFile)
			if err != nil {
				log.Fatal(err)
			}

			var reader io.ReadCloser
			fileAvailable := cmd.Flags().Changed("inputFile")
			urlAvailable := cmd.Flags().Changed("inputUrl")

			if fileAvailable {
				//Load File
				// we explicitly want user input here, hence  #nosec G304
				reader, err = os.Open(inputFile)
				if err != nil {
					log.Fatal(err)
				}
			} else if urlAvailable {
				// Load file from url
				response, err := http.Get(inputUrl)
				if err != nil {
					log.Fatal(err)
				}

				if response.StatusCode < 200 || response.StatusCode > 299 {
					log.Fatal(fmt.Errorf("failed to get file from url: %s", response.Status))
				}

				reader = response.Body
			}

			defer func() {
				if reader != nil {
					if err := reader.Close(); err != nil {
						log.Printf("Error closing file: %s\n", err)
					}
				}
			}()

			users, err := loadAndValidate(reader)
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
			log.Println(fmt.Sprintf("Successfully imported %v users.", len(users)))
		},
	}

	cmd.Flags().StringVar(&configFile, "config", config.DefaultConfigFilePath, "config file")
	cmd.Flags().StringVarP(&inputFile, "inputFile", "i", "", "The json file where the users should be imported from.")
	cmd.Flags().StringVarP(&inputUrl, "inputUrl", "u", "", "The url to a json file where the users should be imported from.")
	cmd.MarkFlagsMutuallyExclusive("inputFile", "inputUrl")
	return cmd
}

// loadAndValidate reads json from an io.Reader so we read every entry separate and validate it. We go through the whole
// array to print out every validation error in the input data.
func loadAndValidate(input io.Reader) ([]ImportOrExportEntry, error) {
	dec := json.NewDecoder(input)

	// read the open bracket
	_, err := dec.Token()
	if err != nil {
		return nil, err
	}

	users := []ImportOrExportEntry{}
	v := validator.New()

	numErrors := 0
	index := 0
	// while the array contains values
	for dec.More() {
		index = index + 1
		var userEntry ImportOrExportEntry
		// decode one ImportEntry
		err := dec.Decode(&userEntry)
		if err != nil {
			errorMsg := fmt.Sprintf("Error at entry %v : %v", index, err.Error())
			log.Println(errorMsg)
			return nil, err
		}

		if err := userEntry.validate(v); err != nil {
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
		return nil, errors.New(errMsg)
	}

	return users, nil
}

// commits the list of ImportEntries to the database. Wrapped in a transaction so if something fails no new users are added.
func addToDatabase(entries []ImportOrExportEntry, persister persistence.Persister) error {
	tx := persister.GetConnection()
	err := tx.Transaction(func(tx *pop.Connection) error {
		importer := Importer{
			persister:       persister,
			tx:              tx,
			importTimestamp: time.Now().UTC(),
		}
		for i, v := range entries {
			userModel, err := importer.createUser(v)
			if err != nil {
				return fmt.Errorf("failed to create user entry nr. %d: %w", i, err)
			}

			for _, e := range v.Emails {
				emailModel, err := importer.createEmailAddress(userModel.ID, e)
				if err != nil {
					return fmt.Errorf("failed to create email address \"%s\" for user entry nr. %d: %w", e.Address, i, err)
				}
				if e.IsPrimary {
					err = importer.createPrimaryEmailAddress(userModel.ID, emailModel.ID)
					if err != nil {
						return fmt.Errorf("failed to set email \"%s\" as primary for user entry nr. %d: %w", e.Address, i, err)
					}
				}
			}

			if v.Username != nil {
				err = importer.createUsername(userModel.ID, *v.Username)
				if err != nil {
					return fmt.Errorf("failed to create username \"%v\" for user entry nr. %d: %w", v.Username, i, err)
				}
			}

			for _, credential := range v.WebauthnCredentials {
				err = importer.createWebauthnCredential(userModel.ID, credential)
				if err != nil {
					return fmt.Errorf("failed to create webauthn credential \"%s\" for user entry nr. %d: %w", credential.ID, i, err)
				}
			}

			if v.Password != nil {
				err = importer.createPasswordCredential(userModel.ID, *v.Password)
				if err != nil {
					return fmt.Errorf("failed to create password for user entry nr. %d: %w", i, err)
				}
			}

			if v.OTPSecret != nil {
				err = importer.createOTPSecret(userModel.ID, *v.OTPSecret)
				if err != nil {
					return fmt.Errorf("failed to create otp secret for user entry nr. %d: %w", i, err)
				}
			}
		}
		return nil
	})

	return err
}
