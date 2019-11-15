package core

import (
	"net"

	"github.com/squ94wk/mqtt-broker/pkg/action"
	"github.com/squ94wk/mqtt-common/pkg/packet"
)

type PacketHandlerParent struct {
	Error              (error)
	OnNewConnectAction (action.Connect)
}

type PacketHandler struct {
	parent   PacketHandlerParent
	shutdown chan interface{}
	log      zap.Logger
}

type InboundPacketBundle struct {
	packet packet.Packet
	conn   net.Conn
}

var (
	inboundPackets chan InboundPacketBundle
)

func init() {
	inboundPackets := make(chan InboundPacketBundle, 10)
}

func (h PacketHandler) Start() {
	for {
		select {
		case p := <-inboundPackets:
			switch p.(Type) {
			case packet.Connect:
				connect := p.(*Connect)
				connectAction := action.NewConnect(connect, p.conn)

				h.parent.OnNewConnectAction(connectAction)
			}

		case s := <-h.shutdown:
			break
		}
	}
}

func PushInboundPacketBundle(b InboundPacketBundle) {
	inboundPackets <- b
}
