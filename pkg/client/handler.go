package client

import (
	"fmt"

	"github.com/squ94wk/mqtt-broker/pkg/connection"
	"github.com/squ94wk/mqtt-broker/pkg/message"
	"github.com/squ94wk/mqtt-broker/pkg/session"
	"github.com/squ94wk/mqtt-common/pkg/packet"
	"go.uber.org/zap"
)

type parent interface {
	Error(error)
}

type Handler struct {
	parent   parent
	conn     *connection.Connection
	flow     flow
	actions  chan func() error
	shutdown chan struct{}
	log      *zap.Logger
}

type flow interface {
	onPacket(Handler, packet.Packet) (flow, error)
	onError(Handler, error) flow
}

func NewHandler(parent parent, conn *connection.Connection, log *zap.Logger) Handler {
	return Handler{
		parent:   parent,
		conn:     conn,
		flow:     connecting{},
		actions:  make(chan func() error, 4),
		shutdown: make(chan struct{}, 1),
		log:      log,
	}
}

func (h Handler) Start() {
	h.log.Debug("client handler started")
	for {
		select {
		case _ = <-h.shutdown:
			return

		case reading := <-h.conn.Packets():
			pkt, err := reading()
			if err != nil {
				h.log.Error(fmt.Sprintf("failed to handle inbound packet: %v", err))
				//TODO: handle properly
				return
			}
			h.handleInboundPacket(pkt) //flow

		case action := <-h.actions:
			err := action()
			if err != nil {
				h.parent.Error(err)
				return
			}
		}
	}
}

func (h Handler) Deliver(msg message.Message) {
	h.actions <- func() error {
		//TODO: make publish packet
		//h.conn.Deliver(pkt)
		return nil
	}
}

func (h Handler) Disconnect(reason packet.DisconnectReason, reasonMsg string) {
	h.actions <- func() error {
		var disconnect packet.Disconnect
		disconnect.SetDisconnectReason(reason)
		props := packet.NewProperties()
		if reasonMsg != "" {
			props.Add(packet.NewProperty(packet.ReasonString, packet.StringPropPayload(reasonMsg)))
		}
		disconnect.SetProps(props)
		h.conn.DeliverFinally(disconnect)
		err := h.conn.Close()
		if err != nil {
			h.parent.Error(fmt.Errorf("failed to disconnect: %v", err))
		}
		return nil
	}
}

func (h Handler) performConnect(clientId string, cleanStart bool) (string, bool, error) {
	//delegate work to the handler's main goroutine as an action
	h.log.Debug("perform connect", zap.String("clientId", clientId), zap.Bool("clean start", cleanStart))
	s, present, err := session.BindToSession(clientId, cleanStart, &h)
	if err != nil {
		return "", false, err
	}
	return s.ClientID(), present, nil
	// session (clean start or persisted)
	//   assigned client id

	// Connect Action
	// session takeover:
	//   will message
	//   disconnect packet
	// finish
	//   connack
	// post finish
	//   message delivery
	//   keep alive monitoring
}

func (h Handler) connackWithError(err error) {
	var connack packet.Connack
	connack.SetSessionPresent(false)
	connack.SetConnectReason(packet.ConnectUnspecifiedError)
	connack.Props().Add(packet.NewProperty(
		packet.ReasonString,
		packet.StringPropPayload(fmt.Sprintf("%v", err)),
	))
	h.conn.DeliverFinally(connack)
	err = h.conn.Close()
	if err != nil {
		h.parent.Error(fmt.Errorf("failed to close client after error: %v", err))
	}
}

func (h Handler) connackWithSuccess(clientID string, sessionPresent bool) {
	var connack packet.Connack
	connack.SetSessionPresent(sessionPresent)
	//TODO: Add assigned client id as prop
	connack.SetConnectReason(packet.ConnectSuccess)
	h.conn.Deliver(connack)
}

func (h Handler) handleInboundPacket(pkt packet.Packet) {
	nextFlow, err := h.flow.onPacket(h, pkt)
	if err != nil {
		h.parent.Error(fmt.Errorf("failed to handle inbound packet: flow reports error: %v", err))
		return
	}
	h.flow = nextFlow
}
