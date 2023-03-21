package config

import (
	"context"
	"os"
	"time"

	"github.com/sethvargo/go-envconfig"
	"gopkg.in/yaml.v2"
)

type CORSConfig struct {
	CORS struct {
		AllowedOrigins   []string `yaml:"origins" env:"ALLOWED_ORIGINS,overwrite"`
		AllowedMethods   []string `yaml:"methods" env:"ALLOWED_METHODS,overwrite"`
		AllowedHeaders   []string `yaml:"headers" env:"ALLOWED_HEADERS,overwrite"`
		AllowCredentials bool     `yaml:"credentials" env:"ALLOW_CREDENTIALS,overwrite"`
	} `yaml:"cors"`
}

func (cc *CORSConfig) Validate() error {
	return nil
}

func BuildNewCorsConfig(path string) func() (*CORSConfig, error) {
	return func() (*CORSConfig, error) {
		var config CORSConfig
		config.CORS.AllowedOrigins = []string{"*"}
		config.CORS.AllowedMethods = []string{"*"}
		config.CORS.AllowedHeaders = []string{"*"}
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
