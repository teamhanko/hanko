package migrate

import (
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/config"
	"github.com/teamhanko/hanko/persistence"
	"log"
)

func NewMigrateUpCommand(config *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "migrate the database up",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			log.Println("migrate up")
			persister, err := persistence.New(config.Database)
			if err != nil {
				log.Fatal(err)
			}
			err = persister.MigrateUp()
			if err != nil {
				log.Fatal(err)
			}
		},
	}
	return cmd
}
