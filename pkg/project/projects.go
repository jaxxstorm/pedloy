// pkg/project/projects.go

package project

import "fmt"

type StackConfig struct {
	Name       string `yaml:"name"`
	AWSProfile string `yaml:"aws_profile,omitempty"`
}

type Stacks []StackConfig

func (s *Stacks) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Try as []string first
	var names []string
	if err := unmarshal(&names); err == nil {
		*s = make([]StackConfig, len(names))
		for i, n := range names {
			(*s)[i] = StackConfig{Name: n}
		}
		return nil
	}
	// Try as []map
	var configs []StackConfig
	if err := unmarshal(&configs); err == nil {
		*s = configs
		return nil
	}
	return fmt.Errorf("stacks must be a list of strings or a list of stack config objects")
}

type Project struct {
	Name       string   `yaml:"name"`
	Stacks     Stacks   `yaml:"stacks"`
	DependsOn  []string `yaml:"dependsOn"`
	Dir        string   `yaml:"dir,omitempty"`
	AWSProfile string   `yaml:"aws_profile,omitempty"`
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
