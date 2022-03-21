/*
Copyright © 2022 Hanko GmbH <developers@hanko.io>

*/
package cmd

import (
	"fmt"
	"github.com/teamhanko/hanko/config"
	"net/http"

	"github.com/spf13/cobra"
)

func NewServeCommand(config *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the hanko server",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("serving on %v", config.Server.Public.Address)
			err := http.ListenAndServe(config.Server.Public.Address, nil)
			if err != nil {
				return
			}
		},
	}
}
