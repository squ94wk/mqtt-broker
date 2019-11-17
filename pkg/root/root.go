package root

import (
	"fmt"
	"net"

	"github.com/squ94wk/mqtt-broker/pkg/action"
	"github.com/squ94wk/mqtt-broker/pkg/client"
	"github.com/squ94wk/mqtt-broker/pkg/listener"
	"github.com/squ94wk/mqtt-broker/pkg/log"
	"github.com/squ94wk/mqtt-common/pkg/packet"
	"go.uber.org/zap"
)

type Root struct {
	log *zap.Logger
}

func NewBroker() Root {
	return Root{
		log: log.Logger,
	}
}

func (r Root) Start() {
	listener := listener.NewHandler(r, r.log)
	go listener.Start()
	client := client.NewHandler(r, r.log)
	go client.Start()

	select {}
}

func (r Root) Error(err error) {
	r.log.Error(fmt.Sprintf("error occured %v", err))
}

func (r Root) OnNewConnection(conn net.Conn) {
	client.ClientConnected(conn)
}

func (r Root) OnPacket(packet packet.Packet) {
	action.NewPacket(packet)
}
