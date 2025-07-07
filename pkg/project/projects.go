// pkg/project/projects.go

package project

import "fmt"

type StackConfig struct {
	Name string            `yaml:"name"`
	Env  map[string]string `yaml:"env,omitempty"`
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

	// Try as []map[string]interface{} to support env
	var rawConfigs []map[string]interface{}
	if err := unmarshal(&rawConfigs); err == nil {
		configs := make([]StackConfig, len(rawConfigs))
		for i, raw := range rawConfigs {
			// Name is required
			name, ok := raw["name"].(string)
			if !ok {
				return fmt.Errorf("stack config missing required 'name' field")
			}
			configs[i].Name = name

			// Env is optional
			if envRaw, ok := raw["env"]; ok {
				envMap := make(map[string]string)
				switch envTyped := envRaw.(type) {
				case map[interface{}]interface{}:
					for k, v := range envTyped {
						ks, kOk := k.(string)
						vs, vOk := v.(string)
						if kOk && vOk {
							envMap[ks] = vs
						}
					}
				case map[string]interface{}:
					for k, v := range envTyped {
						if vs, vOk := v.(string); vOk {
							envMap[k] = vs
						}
					}
				}
				configs[i].Env = envMap
			}
		}
		*s = configs
		return nil
	}

	// Try as []StackConfig (for backwards compatibility)
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
