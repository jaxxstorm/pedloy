package deploy

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/jaxxstorm/pedloy/pkg/auto"
	"github.com/jaxxstorm/pedloy/pkg/config"
	"github.com/jaxxstorm/pedloy/pkg/project"
	"github.com/jaxxstorm/pedloy/pkg/util"
)

// Command creates the deploy command.
func Command() *cobra.Command {
	v := viper.New()

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy Pulumi stacks",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Bind flags to viper
			v.BindPFlags(cmd.Flags())

			// Load configuration
			projects, err := config.LoadConfig(v)
			if err != nil {
				return fmt.Errorf("error loading config: %w", err)
			}

			// Validate dependencies
			if err := util.ValidateDependencies(projects); err != nil {
				return fmt.Errorf("invalid dependencies: %w", err)
			}

			// Set up project source
			source := project.ProjectSource{
				IsGit:     v.GetString("git-url") != "",
				GitURL:    v.GetString("git-url"),
				GitBranch: v.GetString("git-branch"),
				LocalPath: v.GetString("path"),
			}

			org := v.GetString("org")
			jsonLogger := v.GetBool("json")
			preview := v.GetBool("preview")

			// Perform preview or deployment
			if preview {
				err = util.PreviewExecution(projects, "deploy")
				if err != nil {
					return fmt.Errorf("preview failed: %w", err)
				}
			} else {
				errorFile := v.GetString("error-file")
				auto.Deploy(org, projects, source, jsonLogger, errorFile)
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().String("config", "projects.yml", "Path to the configuration file")
	cmd.Flags().String("org", "", "The Pulumi org stacks live in")
	cmd.Flags().String("path", "", "The path to Pulumi projects")
	cmd.Flags().String("git-url", "", "The Git repository URL for projects")
	cmd.Flags().String("git-branch", "main", "The Git branch to use")
	cmd.Flags().Bool("preview", false, "Preview the deployment plan")
	cmd.Flags().Bool("json", false, "Enable JSON logging")
	cmd.Flags().String("error-file", "", "Path to error log file (optional)")

	return cmd
}
