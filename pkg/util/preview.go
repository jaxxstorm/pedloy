package util

import (
	"fmt"

	"github.com/jaxxstorm/pedloy/pkg/graph"
	"github.com/jaxxstorm/pedloy/pkg/project"
)

// PreviewExecution provides a preview of the execution plan.
func PreviewExecution(projects []project.Project, mode string) error {
	// Get execution groups
	executionGroups, err := graph.GetExecutionGroups(projects)
	if err != nil {
		return fmt.Errorf("failed to determine execution groups: %w", err)
	}

	// Print the execution groups for preview
	fmt.Printf("\n%s Plan:\n", mode)
	for i, group := range executionGroups {
		fmt.Printf("Stage %d:\n", i+1)
		for _, stack := range group {
			fmt.Printf("  %s\n", stack)
		}
		fmt.Println()
	}

	return nil
}
