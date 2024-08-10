package migrate

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/persistence"
	"log"
)

func NewMigrateUpCommand() *cobra.Command {
	var (
		configFile string
	)

	cmd := &cobra.Command{
		Use:   "up",
		Short: "migrate the database up",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			log.Println("migrate up")

			cfg, err := config.Load(&configFile)
			if err != nil {
				log.Fatal(err)
			}

			persister, err := persistence.New(cfg.Database)
			if err != nil {
				log.Fatal(err)
			}
			err = persister.MigrateUp()
			if err != nil {
				log.Fatal(err)
			}

			err = persister.GetConnection().Close()
			if err != nil {
				log.Println(fmt.Errorf("failed to close db connection: %w", err))
			}
		},
	}

	cmd.Flags().StringVar(&configFile, "config", config.DefaultConfigFilePath, "config file")

	return cmd
}
