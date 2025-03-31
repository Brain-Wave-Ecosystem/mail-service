package config

import "github.com/Brain-Wave-Ecosystem/go-common/pkg/config"

type Config struct {
	config.DefaultServiceConfig
	Rabbit RabbitConfig `envPrefix:"RABBIT_"`
}

type RabbitConfig struct {
	URL string `env:"URL"`
}
