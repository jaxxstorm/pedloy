// cmd/pedloy/main.go

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/jaxxstorm/pedloy/cmd/pedloy/deploy"
	"github.com/jaxxstorm/pedloy/cmd/pedloy/destroy"
	"github.com/jaxxstorm/pedloy/cmd/pedloy/version"
	"github.com/jaxxstorm/pedloy/pkg/contract"
)

var (
	githubToken string
	debug       bool
)

func configureCLI() *cobra.Command {
	v := viper.New()

	rootCommand := &cobra.Command{
		Use:  "pedloy",
		Long: "Deploy Pulumi stacks in order",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Bind flags to viper
			if err := v.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			// Set default values for viper
			v.SetDefault("config", "projects.yml")
			v.SetDefault("preview", false)
			v.SetDefault("json", false)
			return nil
		},
	}

	// Add subcommands
	rootCommand.AddCommand(deploy.Command())
	rootCommand.AddCommand(destroy.Command())
	rootCommand.AddCommand(version.Command())

	// Persistent Flags
	rootCommand.PersistentFlags().Bool("preview", false, "Preview the order of operations.")
	rootCommand.PersistentFlags().Bool("json", false, "Output all logs as JSON.")
	rootCommand.PersistentFlags().String("org", "", "The Pulumi org stacks live in.")
	rootCommand.PersistentFlags().String("path", "", "The path the Pulumi projects live in.")
	rootCommand.PersistentFlags().String("config", "projects.yml", "The projects.yml file to read.")

	return rootCommand
}

func main() {
	rootCommand := configureCLI()

	if err := rootCommand.Execute(); err != nil {
		contract.IgnoreIoError(fmt.Fprintf(os.Stderr, "%s", err))
		os.Exit(1)
	}
}
