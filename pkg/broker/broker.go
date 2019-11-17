package broker

import (
	"fmt"
	"net"

	"github.com/squ94wk/mqtt-broker/pkg/action"
	"github.com/squ94wk/mqtt-broker/pkg/client"
	"github.com/squ94wk/mqtt-broker/pkg/connect"
	"github.com/squ94wk/mqtt-broker/pkg/listener"
	"github.com/squ94wk/mqtt-broker/pkg/log"
	"github.com/squ94wk/mqtt-common/pkg/packet"
	"go.uber.org/zap"
)

type Broker struct {
	log *zap.Logger
}

func NewBroker() Broker {
	return Broker{
		log: log.Logger,
	}
}

func (r Broker) Start() {
	listener := listener.NewHandler(&r, r.log)
	go listener.Start()

	client := client.NewHandler(&r, r.log)
	go client.Start()

	action := action.NewHandler(&r, r.log)
	go action.Start()

	select {}
}

func (r *Broker) Error(err error) {
	r.log.Error(fmt.Sprintf("error occured %v", err))
}

func (r *Broker) OnNewConnection(conn net.Conn) {
	r.log.Info("new connection")
	client.ClientConnected(conn)
}

func (r *Broker) OnPacket(packet packet.Packet) {
	r.log.Info("new packet")
	action.NewPacket(packet)
}

func (r *Broker) OnConnectAction(action connect.Action) {
	r.log.Info("new connect action")
}
