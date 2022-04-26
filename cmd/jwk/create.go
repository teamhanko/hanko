package jwk

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/config"
	"github.com/teamhanko/hanko/crypto/jwk"
	"github.com/teamhanko/hanko/persistence"
)

func NewCreateCommand(cfg *config.Config, persister persistence.Persister) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create JSON Web Keys and persist them in the Database",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("create called")
			jwkPersister := persister.GetJwkPersister()
			_, err := jwk.NewDefaultManager(cfg.Secrets.Keys, jwkPersister)
			if err != nil {
				fmt.Println(err)
				return
			}
		},
	}
	return cmd
}
