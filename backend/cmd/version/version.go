package version

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/build_info"
)

func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version and exit",
		Long: `Prints the version of the hanko binary.
For all non 'clean' semver tags (e.g. vX.Y.Z) the format is the following: vX.Y.Z-CC-CH[-dirty].
vX.Y.Z: the last tagged semver tag
CC: Commits since the last tag
CH: The commit short hash of the current commit
[-dirty]: is appended if there are any changes that are not commited yet`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(build_info.GetVersion())
		},
	}
}

func RegisterCommands(parent *cobra.Command) {
	cmd := NewVersionCommand()
	parent.AddCommand(cmd)
}
