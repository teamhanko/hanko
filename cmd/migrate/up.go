package migrate

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/persistence"
	"os"
)

func NewMigrateUpCommand(persister *persistence.Persister) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "migrate the database up",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			err := persister.MigrateUp()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}
	return cmd
}
