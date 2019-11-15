package main

import "github.com/squ94wk/mqtt-broker/pkg/broker"

func main() {
	broker := broker.Broker{}

	broker.Start()
}
