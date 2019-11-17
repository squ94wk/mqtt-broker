package action

import (
	"github.com/squ94wk/mqtt-broker/pkg/connect"
	"github.com/squ94wk/mqtt-common/pkg/packet"
)

type Parent interface {
	Error(error)
	OnConnectAction(connect.Action)
}

type Handler struct {
	parent   Parent
	shutdown chan bool
}

var (
	actions chan func(Handler) error
)

func init() {
	actions = make(chan func(Handler) error, 16)
}

func (h Handler) Start() {
	for {
		select {
		case action := <-actions:
			err := action(h)
			if err != nil {
				h.parent.Error(err)
			}

		case _ = <-h.shutdown:
			break
		}
	}
}

func NewPacket(p packet.Packet) {
	actions <- func(h Handler) error {
		switch p.(type) {
		case *packet.Connect:
			//connectPkt := p.(*packet.Connect)
			_ = p.(*packet.Connect)
			action := connect.NewAction() // pass conn pkt contents
			h.parent.OnConnectAction(action)
		default:
			break
		}
		return nil
	}
}
