package client

import (
	"net"

	"github.com/squ94wk/mqtt-common/pkg/packet"
	"go.uber.org/zap"
)

type Parent interface {
	Error(error)
}

type Handler struct {
	parent   Parent
	log      *zap.Logger
	shutdown chan bool
}

var (
	actions chan func(*Handler) error
)

func NewHandler(parent Parent, log *zap.Logger) Handler {
	return Handler{
		parent:   parent,
		log:      log,
		shutdown: make(chan bool, 1),
	}
}

func (h *Handler) Start() {
	for {
		select {
		case _ = <-h.shutdown:
			break
		case action := <-actions:
			err := action(h)
			if err != nil {
				h.parent.Error(err)
			}
		}
	}
}

func ClientConnected(conn net.Conn) {
	actions <- func(h *Handler) error {
		client := NewClient(conn, h)
		go client.Start()
		return nil
	}
}

func (h Handler) Error(err error) {
	h.parent.Error(err)
}

func (h Handler) OnPacket(packet.Packet) {

}
