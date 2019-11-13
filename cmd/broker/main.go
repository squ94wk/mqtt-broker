package main

import "github.com/squ94wk/mqtt-broker/pkg/actor"

func main() {
	root := actor.Root{}

	root.Start()
}
