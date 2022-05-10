package jwk

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/config"
	"github.com/teamhanko/hanko/crypto/jwk"
	"github.com/teamhanko/hanko/persistence"
	"log"
)

func NewCreateCommand(cfg *config.Config, persister persistence.Persister) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create JSON Web Key and print them in the console",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("create called")
			generator := jwk.RSAKeyGenerator{}
			key, err := generator.Generate("key1")
			if err != nil {
				log.Panicln(err)
			}
			j, err := json.Marshal(key)
			if err != nil {
				log.Panicln(err)
			}
			fmt.Println(string(j))
		},
	}
	return cmd
}
