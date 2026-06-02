package user

import "github.com/spf13/cobra"

func NewConvertCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "convert",
		Short: "Utilities for converting external user data to Hanko import data",
		Long:  ``,
	}

	cmd.AddCommand(NewFirebaseCommand())

	return cmd
}
