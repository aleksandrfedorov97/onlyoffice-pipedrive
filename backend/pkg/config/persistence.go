package config

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/sethvargo/go-envconfig"
	"gopkg.in/yaml.v2"
)

type PersistenceConfig struct {
	Persistence struct {
		URL string `yaml:"url" env:"PERSISTENCE_URL,overwrite"`
	} `yaml:"persistence"`
}

func (p *PersistenceConfig) Validate() error {
	p.Persistence.URL = strings.TrimSpace(p.Persistence.URL)
	return nil
}

func BuildNewPersistenceConfig(path string) func() (*PersistenceConfig, error) {
	return func() (*PersistenceConfig, error) {
		var config PersistenceConfig
		if path != "" {
			file, err := os.Open(path)
			if err != nil {
				return nil, err
			}
			defer file.Close()

			decoder := yaml.NewDecoder(file)

			if err := decoder.Decode(&config); err != nil {
				return nil, err
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()
		if err := envconfig.Process(ctx, &config); err != nil {
			return nil, err
		}

		return &config, config.Validate()
	}
}
