package jwt

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/config"
	"github.com/teamhanko/hanko/crypto/jwk"
	"github.com/teamhanko/hanko/persistence"
	"github.com/teamhanko/hanko/session"
	"log"
)

func NewCreateCommand(config *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [user_id]",
		Short: "generate a JSON Web Token for a given user_id",
		Long:  ``,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("user_id required")
			}
			if _, err := uuid.FromString(args[0]); err != nil {
				return errors.New("user_id is not a uuid")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			persister, err := persistence.New(config.Database)
			if err != nil {
				log.Fatal(err)
			}
			jwkPersister := persister.GetJwkPersister()
			jwkManager, err := jwk.NewDefaultManager(config.Secrets.Keys, jwkPersister)
			if err != nil {
				fmt.Printf("failed to create jwk persister: %s", err)
				return
			}

			sessionManager, err := session.NewManager(jwkManager, config.Session)
			if err != nil {
				fmt.Printf("failed to create session generator: %s", err)
				return
			}

			token, err := sessionManager.GenerateJWT(uuid.FromStringOrNil(args[0]))
			if err != nil {
				fmt.Printf("failed to generate token: %s", err)
				return
			}

			fmt.Printf("token: %s", token)
		},
	}
	return cmd
}
