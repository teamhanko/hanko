/*
Copyright © 2022 Hanko GmbH <developers@hanko.io>

*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/cmd/migrate"
	"github.com/teamhanko/hanko/config"
	"github.com/teamhanko/hanko/persistence"
	"os"
)

var (
	cfgFile   string
	cfg       *config.Config
	persister *persistence.Persister
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "hanko",
	}
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	initConfig()
	initPersister()
	migrate.RegisterCommands(cmd, persister)
	cmd.AddCommand(NewServeCommand(cfg))

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
