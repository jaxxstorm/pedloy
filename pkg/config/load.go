// pkg/config/load.go - Load the configuration file and return a slice of projects
package config

import (
	"fmt"
	"github.com/jaxxstorm/pedloy/pkg/project"
	"gopkg.in/yaml.v3"
	"os"

	"github.com/spf13/viper"
)

func LoadConfig(v *viper.Viper) ([]project.Project, error) {
	configPath := v.GetString("config") // Use viper to get the config path
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var cfg project.Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg.Projects, nil
}
