package main

import (
	"github.com/squ94wk/mqtt-broker/pkg/broker"
)

func main() {
	//conf := config.ParseConfig()
	broker := broker.NewBroker()
	broker.Start()
}
