package client

type Config struct {
	address string
}

func NewConfig(options ...func(*Config)) Config {
	var config Config
	for _, option := range options {
		option(&config)
	}

	return config
}

func ListenAddress(address string) func(*Config) {
	return func(h *Config) {
		h.address = address
	}
}
