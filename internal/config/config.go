package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	RabbitmqUrl string `required:"true" split_words:"true"`
}

func New() (*Config, error) {
	cfg := new(Config)
	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, fmt.Errorf("failed loading config %w", err)
	}

	return cfg, nil
}
