/*
Copyright © 2022 Hanko GmbH <developers@hanko.io>
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/cmd/isready"
	"github.com/teamhanko/hanko/backend/cmd/jwk"
	"github.com/teamhanko/hanko/backend/cmd/jwt"
	"github.com/teamhanko/hanko/backend/cmd/migrate"
	"github.com/teamhanko/hanko/backend/cmd/serve"
	"github.com/teamhanko/hanko/backend/cmd/version"
	"log"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "hanko",
	}

	migrate.RegisterCommands(cmd)
	serve.RegisterCommands(cmd)
	isready.RegisterCommands(cmd)
	jwk.RegisterCommands(cmd)
	jwt.RegisterCommands(cmd)
	version.RegisterCommands(cmd)

	return cmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cmd := NewRootCmd()

	err := cmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
