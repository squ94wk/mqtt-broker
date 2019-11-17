package root

type BrokerConfig struct {
	address string
}

func NewConfig(options ...func(*BrokerConfig)) BrokerConfig {
	var config BrokerConfig
	for _, option := range options {
		option(&config)
	}

	return config
}

func ListenAddress(address string) func(*BrokerConfig) {
	return func(h *BrokerConfig) {
		h.address = address
	}
}
