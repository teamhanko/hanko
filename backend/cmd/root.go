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
