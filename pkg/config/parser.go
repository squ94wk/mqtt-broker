package config

import (
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

func Parse() Config {
	var config Config
	fromFile(&config)
	fromEnv(&config)

	return config
}

func fromFile(c *Config) {
	f, err := os.Open("config.yml")
	if err != nil {
		fmt.Println("failed to read config file, or none there")
		return
	}

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(c)
	if err != nil {
		fmt.Println("failed to read config file")
	}
}

func fromEnv(c *Config) {
	err := envconfig.Process("", c)
	if err != nil {
		fmt.Println("failed to read config file")
	}
}
