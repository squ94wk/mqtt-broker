package config

import (
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

type Client struct {
	address string `yaml:"address", envconfig:"SERVER_ADDRESS"`
}

type Server struct {
	Client
}

type Config struct {
	Server struct {
	} `yaml:"server"`
}

func Parse() Config {
	var config Config
	fromFile(&config)
	fromEnv(&config)

	return config
}

func defaults() Config {
	var config Config
	config.Server.address = "localhost:1883"

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