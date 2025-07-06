package version

import (
	"fmt"

	"github.com/jaxxstorm/pedloy/pkg/version"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	command := &cobra.Command{
		Use:   "version",
		Short: "Get the current version",
		Long:  `Get the current version of pedloy`,
		RunE: func(*cobra.Command, []string) error {
			v := version.GetVersion()
			fmt.Println(v)
			return nil
		},
	}
	return command
}
