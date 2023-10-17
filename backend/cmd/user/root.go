package user

import (
	"github.com/spf13/cobra"
)

func NewUserCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "user",
		Short: "User import/export tools",
		Long:  `Add the ability to import/export users into/from the hanko database.`,
	}
}

func RegisterCommands(parent *cobra.Command) {
	command := NewUserCommand()
	parent.AddCommand(command)
	command.AddCommand(NewImportCommand())
	command.AddCommand(NewGenerateCommand())
	command.AddCommand(NewExportCommand())
}
