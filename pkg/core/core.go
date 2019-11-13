package core

import "net"

type HandlerParent struct {
	Error             (error)
	OnNewConnectAtion (action.Connect)
}

type Handler struct {
	parent   HandlerParent
	shutdown chan interface{}
	log      zap.Production
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

func (h Handler) Start() {
	for {
		select {
		case p := <-inboundPackets:
			switch p.(Type) {
			case packet.Connect:
				connect := p.(*Connect)
				connectAction := action.NewConnect(connect, p.conn)

				h.parent.OnNewConnectAtion(connectAction)
			}

		case s := <-h.shutdown:
			break
		}
	}
}

func PushInboundPacketBundle(b InboundPacketBundle) {
	inboundPackets <- b
}
