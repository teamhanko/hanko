/*
Copyright © 2022 Hanko GmbH <developers@hanko.io>
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/v2/cmd/cleanup"
	"github.com/teamhanko/hanko/backend/v2/cmd/isready"
	"github.com/teamhanko/hanko/backend/v2/cmd/jwk"
	"github.com/teamhanko/hanko/backend/v2/cmd/jwt"
	"github.com/teamhanko/hanko/backend/v2/cmd/migrate"
	"github.com/teamhanko/hanko/backend/v2/cmd/schema"
	"github.com/teamhanko/hanko/backend/v2/cmd/serve"
	"github.com/teamhanko/hanko/backend/v2/cmd/siwa"
	"github.com/teamhanko/hanko/backend/v2/cmd/user"
	"github.com/teamhanko/hanko/backend/v2/cmd/version"
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
	user.RegisterCommands(cmd)
	siwa.RegisterCommands(cmd)
	schema.RegisterCommands(cmd)
	cleanup.RegisterCommands(cmd)

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
