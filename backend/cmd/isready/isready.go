package isready

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/config"
	"log"
	"net"
	"net/http"
)

func NewIsReadyCommand(config *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "isready",
		Args: cobra.OnlyValidArgs,
		ValidArgs: []string{"admin", "public"},
		Short: "Health check a service",
		Long: `Checks if the specified service is healthy. Possible values are "admin" and "public".
Uses the "/health/ready" endpoint to check if the service is ready to serve requests.
		`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				log.Fatalf("Please specify a service to check")
			}
			service := args[0]
			address := ""
			if service == "admin" {
				address = config.Server.Admin.Address
			}
			if service == "public" {
				address = config.Server.Public.Address
			}
			host, port, err := net.SplitHostPort(address)
			if err != nil {
				log.Fatalf("Could not parse address %s", address)
			}
			if host == "" {
				host = "localhost"
			}
			requestUrl := fmt.Sprintf("http://%s:%s/health/ready", host, port)
			res, err := http.Get(requestUrl)
			if err != nil {
				log.Fatalf("Service %s is not ready", service)
			} else {
				if res.StatusCode != 200 {
					log.Fatalf("Service %s is not ready", service)
				} else {
					log.Println(fmt.Sprintf("Service %s is ready", service))
				}
			}
		},
	}
}

func RegisterCommands(parent *cobra.Command, config *config.Config) {
	cmd := NewIsReadyCommand(config)
	parent.AddCommand(cmd)
}
