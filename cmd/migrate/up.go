package migrate

import (
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/persistence"
	"log"
)

func NewMigrateUpCommand(persister persistence.Migrator) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "migrate the database up",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			err := persister.MigrateUp()
			if err != nil {
				log.Fatal(err)
			}
		},
	}
	return cmd
}
