package jwk

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/v2/crypto/jwk/local_db"

	"log"
)

func NewCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create JSON Web Key and print them in the console",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("create called")
			generator := local_db.RSAKeyGenerator{}
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
