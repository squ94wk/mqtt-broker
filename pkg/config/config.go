package config

type Client struct {
	address string `yaml:"address", envconfig:"BROKER_ADDRESS"`
}

type Broker struct {
	Client Client `yaml:"client"`
}

type Config struct {
	Broker struct {
	} `yaml:"broker"`
}
