package schema

import (
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/cmd/user"
	"log"
)

func NewGenerateImportCommand() *cobra.Command {
	var (
		extractComments bool
		commentPaths    []string
		outPath         string
		doNotReference  bool
	)

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Generate JSON schema for the user import/export",
		Run: func(cmd *cobra.Command, args []string) {
			err := generateSchema(&user.ImportOrExportList{}, outPath, extractComments, doNotReference, commentPaths...)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	cmd.Flags().BoolVarP(&extractComments, "extract_comments", "e", true, "Extract Go comments")
	cmd.Flags().StringSliceVarP(&commentPaths, "comment_paths", "c", []string{"./config", "./ee", "./webhooks"}, "Path to Go sources to extract comments from")
	cmd.Flags().StringVarP(&outPath, "out_path", "o", "./json_schema/hanko.config.json", "Output destination")
	cmd.Flags().BoolVarP(&doNotReference, "no_reference", "d", false, "Do not reference")

	return cmd
}
