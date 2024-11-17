// pkg/project/projects.go
package project

type Project struct {
	Name      string   `yaml:"name"`
	Stacks    []string `yaml:"stacks"`
	DependsOn []string `yaml:"dependsOn"`
}

type Config struct {
	Projects []Project `yaml:"projects"`
}

type ProjectSource struct {
	IsGit     bool
	GitURL    string
	GitBranch string
	LocalPath string
}