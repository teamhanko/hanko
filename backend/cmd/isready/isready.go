package isready

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/config"
	"net/http"
	"os"
)

func NewIsReadyCommand(config *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "isready",
		Args: cobra.OnlyValidArgs,
		ValidArgs: []string{"admin", "public"},
		Short: "Health check a service",
		Long: `Checks if the specified service is healthy. Possible values are "admin" and "public".`,
		Run: func(cmd *cobra.Command, args []string) {
			// check that there is only one argument
			if len(args) != 1 {
				fmt.Println("Please specify a service to check")
				os.Exit(1)
			}
			service := args[0]
			address := ""
			if service == "admin" {
				address = config.Server.Admin.Address
			}
			if service == "public" {
				address = config.Server.Public.Address
			}
			if address[0] == ':' {
				address = "0.0.0.0" + address
			}
			address = "http://" + address
			res, err := http.Get(address + "/health/ready")
			if err != nil {
				fmt.Printf("Service %s is not ready", service)
				os.Exit(1)
			} else {
				if res.StatusCode != 200 {
					fmt.Printf("Service %s is not ready", service)
					os.Exit(1)
				} else {
					fmt.Printf("Service %s is ready", service)
				}
			}
		},
	}
}

func RegisterCommands(parent *cobra.Command, config *config.Config) {
	cmd := NewIsReadyCommand(config)
	parent.AddCommand(cmd)
}
