// pkg/graph/graph.go
package graph

import(
	"fmt"
	"sort"
	"github.com/dominikbraun/graph"
	p "github.com/jaxxstorm/pedloy/pkg/project"
)

// Creates a unique vertex ID for a project and stack combination
func vertexID(project, stack string) string {
	return fmt.Sprintf("%s:%s", project, stack)
}

func containsStack(stacks []string, stack string) bool {
	for _, s := range stacks {
		if s == stack {
			return true
		}
	}
	return false
}

func GetExecutionGroups(projects []p.Project) ([][]string, error) {
	// Create a directed graph
	g := graph.New(graph.StringHash, graph.Directed())

	// Track dependencies for each vertex
	dependencies := make(map[string][]string)

	// First, create a map of all valid project-stack combinations and merge duplicate projects
	validStacks := make(map[string][]string)
	projectDeps := make(map[string][]string) // Track merged dependencies

	for _, project := range projects {
		// Merge stacks for duplicate projects
		if existing, ok := validStacks[project.Name]; ok {
			// Create a map for unique stacks
			stackMap := make(map[string]bool)
			for _, s := range existing {
				stackMap[s] = true
			}
			for _, s := range project.Stacks {
				stackMap[s] = true
			}

			// Convert back to slice
			var mergedStacks []string
			for s := range stackMap {
				mergedStacks = append(mergedStacks, s)
			}
			validStacks[project.Name] = mergedStacks

			// Merge dependencies
			depMap := make(map[string]bool)
			for _, dep := range projectDeps[project.Name] {
				depMap[dep] = true
			}
			for _, dep := range project.DependsOn {
				depMap[dep] = true
			}

			var mergedDeps []string
			for dep := range depMap {
				mergedDeps = append(mergedDeps, dep)
			}
			projectDeps[project.Name] = mergedDeps
		} else {
			validStacks[project.Name] = project.Stacks
			projectDeps[project.Name] = project.DependsOn
		}
	}

	// Add all vertices first (project:stack combinations)
	for projectName, stacks := range validStacks {
		for _, stack := range stacks {
			vertex := vertexID(projectName, stack)
			if err := g.AddVertex(vertex); err != nil {
				return nil, fmt.Errorf("failed to add vertex %s: %w", vertex, err)
			}
		}
	}

	// Add edges for dependencies
	for projectName, stacks := range validStacks {
		deps := projectDeps[projectName]
		for _, stack := range stacks {
			currentVertex := vertexID(projectName, stack)
			dependencies[currentVertex] = []string{}

			// Add edges for each dependency, but only if the dependency exists in the same stack
			for _, dep := range deps {
				// Check if the dependency exists in this stack
				if containsStack(validStacks[dep], stack) {
					depVertex := vertexID(dep, stack)
					if err := g.AddEdge(depVertex, currentVertex); err != nil {
						return nil, fmt.Errorf("failed to add edge from %s to %s: %w", depVertex, currentVertex, err)
					}
					dependencies[currentVertex] = append(dependencies[currentVertex], depVertex)
				}
			}
		}
	}

	// Get vertices in topological order
	order, err := graph.TopologicalSort(g)
	if err != nil {
		return nil, fmt.Errorf("failed to perform topological sort: %w", err)
	}

	// Create concurrent execution groups
	var executionGroups [][]string
	processed := make(map[string]bool)

	// Process all vertices
	for len(processed) < len(order) {
		var currentGroup []string

		// Find all vertices that can be executed
		for _, vertex := range order {
			if processed[vertex] {
				continue
			}

			// Check if all dependencies are processed
			canExecute := true
			for _, dep := range dependencies[vertex] {
				if !processed[dep] {
					canExecute = false
					break
				}
			}

			if canExecute {
				currentGroup = append(currentGroup, vertex)
			}
		}

		// Sort the group for consistent output
		sort.Strings(currentGroup)

		// Mark all vertices in current group as processed
		for _, vertex := range currentGroup {
			processed[vertex] = true
		}

		executionGroups = append(executionGroups, currentGroup)
	}

	return executionGroups, nil
}