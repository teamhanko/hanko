/*
Copyright © 2022 Hanko GmbH <developers@hanko.io>

*/
package cmd

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the hanko server",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("serving on %v", C.Server.Public.Adress)
		err := http.ListenAndServe(C.Server.Public.Adress, nil)
		if err != nil {
			return 
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
