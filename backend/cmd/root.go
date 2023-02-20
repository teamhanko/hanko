/*
Copyright © 2022 Hanko GmbH <developers@hanko.io>
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/cmd/jwk"
	"github.com/teamhanko/hanko/backend/cmd/jwt"
	"github.com/teamhanko/hanko/backend/cmd/migrate"
	"github.com/teamhanko/hanko/backend/cmd/serve"
	"github.com/teamhanko/hanko/backend/cmd/version"
	"github.com/teamhanko/hanko/backend/config"
	"log"
)

var (
	cfgFile string
	cfg     config.Config
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "hanko",
	}

	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	migrate.RegisterCommands(cmd, &cfg)
	serve.RegisterCommands(cmd, &cfg)
	jwk.RegisterCommands(cmd)
	jwt.RegisterCommands(cmd, &cfg)
	version.RegisterCommands(cmd)

	return cmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.OnInitialize(initConfig)
	cmd := NewRootCmd()

	err := cmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

func initConfig() {
	var err error
	conf, err := config.Load(&cfgFile)
	if err != nil {
		log.Fatalf("failed to load config: %s", err)
	}
	if err = conf.Validate(); err != nil {
		log.Fatalf("failed to validate config: %s", err)
	}
	cfg = *conf
}
