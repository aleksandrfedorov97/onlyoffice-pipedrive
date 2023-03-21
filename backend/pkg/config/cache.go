package config

import (
	"context"
	"os"
	"time"

	"github.com/sethvargo/go-envconfig"
	"gopkg.in/yaml.v2"
)

type CacheConfig struct {
	Cache struct {
		Type     int    `yaml:"type" env:"CACHE_TYPE,overwrite"`
		Size     int    `yaml:"size" env:"CACHE_SIZE,overwrite"`
		Address  string `yaml:"address" env:"CACHE_ADDRESS,overwrite"`
		Password string `yaml:"password" env:"CACHE_PASSWORD,overwrite"`
	} `yaml:"cache"`
}

func (b *CacheConfig) Validate() error {
	switch b.Cache.Type {
	case 2:
		if b.Cache.Address == "" {
			return &InvalidConfigurationParameterError{
				Parameter: "Address",
				Reason:    "Redis cache must have a valid address",
			}
		}
		return nil
	default:
		return nil
	}
}

func BuildNewCacheConfig(path string) func() (*CacheConfig, error) {
	return func() (*CacheConfig, error) {
		var config CacheConfig
		config.Cache.Size = 10
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
