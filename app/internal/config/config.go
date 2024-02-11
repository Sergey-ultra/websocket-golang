package config

import (
	"os"
	"sync"
)

type RabbitConfig struct {
	Host     string `yaml:"host" env-default:"7079"`
	Port     string `yaml:"port" env-default:"0.0.0.0"`
	Username string `yaml:"username" env-default:"guest"`
	Password string `yaml:"password" env-default:"guest"`
}

type Config struct {
	Rabbit RabbitConfig
}

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		//logger := logging.GetLogger()
		//logger.Info("read application config")

		var rabbitConfig = RabbitConfig{
			Host:     os.Getenv("RABBIT_HOST"),
			Port:     os.Getenv("RABBIT_PORT"),
			Username: os.Getenv("RABBIT_USERNAME"),
			Password: os.Getenv("RABBIT_PASSWORD"),
		}

		instance = &Config{
			Rabbit: rabbitConfig,
		}

	})
	return instance
}
