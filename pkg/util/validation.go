package util

import (
	"fmt"

	"github.com/jaxxstorm/pedloy/pkg/project"
)

// ValidateDependencies checks for missing dependencies in the project configuration.
func ValidateDependencies(projects []project.Project) error {
	// Build a set of valid project names
	projectNames := make(map[string]struct{})
	for _, project := range projects {
		projectNames[project.Name] = struct{}{}
	}

	// Check dependencies for each project
	for _, project := range projects {
		for _, dep := range project.DependsOn {
			if _, exists := projectNames[dep]; !exists {
				return fmt.Errorf("project %q depends on missing project %q", project.Name, dep)
			}
		}
	}

	return nil
}
