package migrate

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/persistence"
	"log"
	"strconv"
)

var steps int

func NewMigrateDownCommand() *cobra.Command {
	var (
		configFile string
	)

	cmd := &cobra.Command{
		Use:   "down",
		Short: "migrate the database down - given the number of steps",
		Long:  ``,
		Args: func(cmd *cobra.Command, args []string) error {
			var err error
			if steps, err = strconv.Atoi(args[0]); err != nil {
				return err
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("migrate down called")

			cfg, err := config.Load(&configFile)
			if err != nil {
				log.Fatal(err)
			}

			persister, err := persistence.New(cfg.Database)
			if err != nil {
				log.Fatal(err)
			}
			err = persister.MigrateDown(steps)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	cmd.Flags().StringVar(&configFile, "config", config.DefaultConfigFilePath, "config file")

	return cmd
}
