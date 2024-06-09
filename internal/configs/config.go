package configs

import (
	"fmt"
	"os"

	"github.com/nhtuan0700/GoLoad/configs"
	"gopkg.in/yaml.v2"
)

type ConfigFilePath string

type Config struct {
	Log      Log      `yaml:"log"`
	Database Database `yaml:"database"`
	Auth     Auth     `yaml:"auth"`
	GRPC     GRPC     `yaml:"grpc"`
	HTTP     HTTP     `yaml:"http"`
}

func NewConfig(filepath ConfigFilePath) (Config, error) {
	var (
		configBytes = configs.DefaultConfigBytes
		config      = Config{}
		err         error
	)

	if filepath != "" {
		configBytes, err = os.ReadFile(string(filepath))
		if err != nil {
			return Config{}, fmt.Errorf("failed to read YAML file: %w", err)
		}
	}

	err = yaml.Unmarshal(configBytes, &config)
	if err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return config, nil
}
