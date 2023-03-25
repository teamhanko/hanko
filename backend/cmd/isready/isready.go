package isready

import (
	"fmt"
	"github.com/spf13/cobra"
)

func NewIsReadyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "isready",
		Args: cobra.OnlyValidArgs,
		ValidArgs: []string{"admin", "public"},
		Short: "Health check a service",
		Long: `Checks if the specified service is healthy. Possible values are "admin" and "public".`,
		Run: func(cmd *cobra.Command, args []string) {
			// log the args
			fmt.Println(args)
		},
	}
}

func RegisterCommands(parent *cobra.Command) {
	cmd := NewIsReadyCommand()
	parent.AddCommand(cmd)
}
