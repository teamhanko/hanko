package migrate

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/persistence"
	"strconv"
)

var steps int

func NewMigrateDownCommand(persister *persistence.Persister) *cobra.Command {
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
			err := persister.MigrateDown(steps)
			if err != nil {
				fmt.Println(err)
			}
		},
	}
	return cmd
}
