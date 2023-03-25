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
			// check that there is only one argument
			if len(args) != 1 {
				fmt.Println("Please specify a service to check")
				return
			}
			service := args[0]
			fmt.Printf("Service %s is ready", service)
		},
	}
}

func RegisterCommands(parent *cobra.Command) {
	cmd := NewIsReadyCommand()
	parent.AddCommand(cmd)
}
