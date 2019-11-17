package client

import (
	"net"

	"github.com/squ94wk/mqtt-broker/pkg/client/conn"
	"github.com/squ94wk/mqtt-broker/pkg/client/listener"
	"github.com/squ94wk/mqtt-common/pkg/packet"
	"go.uber.org/zap"
)

type Parent interface {
	Error(error)
}

type Handler struct {
	parent Parent
	log    *zap.Logger
}

var (
	shutdown chan bool
)

func init() {
	shutdown = make(chan bool, 1)
}

func NewHandler(parent Parent, log *zap.Logger) Handler {
	return Handler{
		parent: parent,
		log:    log,
	}
}

func (h Handler) Start() {
	listener := listener.NewHandler(
		h,
		h.log,
	)
	listener.Start()
}

func (h Handler) Error(err error) {
	h.parent.Error(err)
}

func (h Handler) OnNewConnection(connection net.Conn) {
	client := conn.NewHandler(connection, h)
	go client.Start()
}

func (h Handler) OnPacket(packet.Packet) {

}
