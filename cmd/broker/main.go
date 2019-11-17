package main

import (
	"github.com/squ94wk/mqtt-broker/pkg/config"
	"github.com/squ94wk/mqtt-broker/pkg/root"
)

func main() {
	conf := config.ParseConfig()
	broker := root.NewBroker(
		conf,
	)
	broker.Start()
}
