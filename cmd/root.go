/*
Copyright © 2022 Hanko GmbH <developers@hanko.io>

*/
package cmd

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/teamhanko/hanko/config"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)


var (
	cfgFile string
 	C *config.Config
	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "hanko",
		Short: "Passwordless Authentication made easy.",
		Long:  `TODO`, //TODO
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Get base path of binary call
		_, b, _, _ := runtime.Caller(0)
		basePath := filepath.Dir(b)

		viper.SetConfigType("yaml")
		viper.AddConfigPath(basePath)
		viper.AddConfigPath("/etc/config")
		viper.AddConfigPath("/etc/secrets")
		viper.AddConfigPath("./config")
		viper.SetConfigName("hanko-config")
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
	C = config.Default()
	err := viper.Unmarshal(C)
	if err != nil {
		panic(fmt.Sprintf("unable to decode config into struct, %v", err))
	}
}
