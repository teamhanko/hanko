package user

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/config"
)

func NewExportCommand() *cobra.Command {
	var (
		configFile string
		outputFile string
	)

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export users from database into a Json file",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			log.Println("Exporting users...")
		},
	}

	cmd.Flags().StringVar(&configFile, "config", config.DefaultConfigFilePath, "config file")
	cmd.Flags().StringVarP(&outputFile, "outputFile", "o", "", "The path of the output file.")
	err := cmd.MarkFlagRequired("outputFile")
	if err != nil {
		log.Println(err)
	}
	return cmd
}
