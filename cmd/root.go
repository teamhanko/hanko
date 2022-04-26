/*
Copyright © 2022 Hanko GmbH <developers@hanko.io>

*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/cmd/jwk"
	"github.com/teamhanko/hanko/cmd/jwt"
	"github.com/teamhanko/hanko/cmd/migrate"
	"github.com/teamhanko/hanko/cmd/serve"
	"github.com/teamhanko/hanko/config"
	"github.com/teamhanko/hanko/persistence"
	"os"
)

var (
	cfgFile   string
	cfg       *config.Config
	persister persistence.Storage
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "hanko",
	}
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	initConfig()
	err := initPersister()
	if err != nil {
		os.Exit(1)
	}
	migrate.RegisterCommands(cmd, persister)
	serve.RegisterCommands(cmd, cfg, persister)
	jwk.RegisterCommands(cmd, cfg, persister)
	jwt.RegisterCommands(cmd, cfg, persister)

	return cmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cmd := NewRootCmd()

	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func initConfig() {
	cfg = config.Load(&cfgFile)
}

func initPersister() error {
	var err error
	persister, err = persistence.New(cfg.Database)
	if err != nil {
		return err
	}
	return nil
}
