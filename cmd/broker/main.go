package main

import (
	"github.com/squ94wk/mqtt-broker/pkg/broker"
)

func main() {
	//conf := config.ParseConfig()
	b := broker.NewBroker()
	b.Start()
}
