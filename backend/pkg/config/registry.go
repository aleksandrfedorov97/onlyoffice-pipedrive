package config

import (
	"context"
	"os"
	"time"

	"github.com/sethvargo/go-envconfig"
	"gopkg.in/yaml.v2"
)

type RegistryConfig struct {
	Registry struct {
		Addresses    []string      `yaml:"addresses" env:"REGISTRY_ADDRESSES,overwrite"`
		CacheTTL     time.Duration `yaml:"cache_duration" env:"REGISTRY_CACHE_DURATION,overwrite"`
		RegistryType int           `yaml:"type" env:"REGISTRY_TYPE,overwrite"`
	} `yaml:"registry"`
}

func (r *RegistryConfig) Validate() error {
	if len(r.Registry.Addresses) <= 0 {
		return &InvalidConfigurationParameterError{
			Parameter: "Addresses",
			Reason:    "Length should be greater than zero",
		}
	}

	return nil
}

func BuildNewRegistryConfig(path string) func() (*RegistryConfig, error) {
	return func() (*RegistryConfig, error) {
		var config RegistryConfig
		config.Registry.CacheTTL = 10 * time.Second
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
