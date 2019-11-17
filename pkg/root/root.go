package root

import (
	"github.com/squ94wk/mqtt-broker/pkg/action"
	"github.com/squ94wk/mqtt-broker/pkg/client"
	"github.com/squ94wk/mqtt-broker/pkg/config"
	"github.com/squ94wk/mqtt-broker/pkg/log"
	"github.com/squ94wk/mqtt-common/pkg/packet"
	"go.uber.org/zap"
)

type Root struct {
	config config.Config
	log    *zap.Logger
}

func NewBroker(config config.Config) Root {
	return Root{
		config: config,
		log:    log.Logger,
	}
}

func (r Root) Start() {
	c := client.NewHandler(
		r,
		client.NewConfig(client.ListenAddress(r.config.address)),
		r.log,
	)
	go c.Start()

	for {
		// idle
	}
}

func (r Root) Error(err error) {
	r.log.Error(err)
}

func (r Root) OnPacket(packet packet.Packet) {
	action.NewPacket(packet)
}
