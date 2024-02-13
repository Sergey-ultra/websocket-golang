package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"sync"
)

type Config struct {
	RabbitConfig struct {
		Host     string `yaml:"host" env-default:"0.0.0.0"`
		Port     string `yaml:"port" env-default:"7079"`
		Username string `yaml:"username" env-default:"guest"`
		Password string `yaml:"password" env-default:"guest"`
	}
}

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		//logger := logging.GetLogger()
		//logger.Info("read application config")
		instance = &Config{}

		if err := cleanenv.ReadConfig("conf.yml", instance); err != nil {
			_, _ = cleanenv.GetDescription(instance, nil)
			panic(err)
			//logger.Info(help)
			//logger.Fatal(err)
		}

	})
	return instance
}
