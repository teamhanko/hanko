package jwk

import (
	"fmt"
	"github.com/spf13/cobra"
)

func NewCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create JSON Web Keys and persist them in the Database",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("create called")
		},
	}
	cmd.Flags().StringP("alg", "a", "RS256", "Which algorithm to use. On of: RS256, ES256, ES512, HS256, EdDSA")
	return cmd
}
