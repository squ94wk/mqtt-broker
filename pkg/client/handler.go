package client

import (
	"fmt"

	"github.com/squ94wk/mqtt-broker/pkg/connection"
	"github.com/squ94wk/mqtt-broker/pkg/message"
	"github.com/squ94wk/mqtt-common/pkg/packet"
	"github.com/squ94wk/mqtt-common/pkg/topic"
	"go.uber.org/zap"
)

type parent interface {
	Error(error)
	PerformConnect(string, bool, Client) (string, bool, error)
	PerformSubscribe(uint16, topic.Filter, byte, bool, bool, byte) (packet.SubackReason, error)
}

type Client struct {
	parent   parent
	conn     *connection.Connection
	state    state
	actions  chan func() error
	shutdown chan struct{}
	log      *zap.Logger
}

type state interface {
	onPacket(Client, packet.Packet) (state, error)
	onError(Client, error) state
}

func NewClient(parent parent, conn *connection.Connection, log *zap.Logger) *Client {
	return &Client{
		parent:   parent,
		conn:     conn,
		state:    initial{},
		actions:  make(chan func() error, 4),
		shutdown: make(chan struct{}, 1),
		log:      log,
	}
}

func (h Client) Start() {
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
			h.handleInboundPacket(pkt) //state

		case action := <-h.actions:
			err := action()
			if err != nil {
				h.parent.Error(err)
				return
			}
		}
	}
}

func (h Client) Deliver(msg message.Message) {
	h.actions <- func() error {
		//TODO: make publish packet
		//h.conn.Deliver(pkt)
		return nil
	}
}

func (h Client) Disconnect(reason packet.DisconnectReason, reasonMsg string) {
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

func (h Client) connackWithError(err error) {
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

func (h Client) connackWithSuccess(clientID string, sessionPresent bool) {
	var connack packet.Connack
	connack.SetSessionPresent(sessionPresent)
	//TODO: Add assigned client id as prop
	connack.SetConnectReason(packet.ConnectSuccess)
	h.conn.Deliver(connack)
}

func (h Client) handleInboundPacket(pkt packet.Packet) {
	nextFlow, err := h.state.onPacket(h, pkt)
	if err != nil {
		h.parent.Error(fmt.Errorf("failed to handle inbound packet: state reports error: %v", err))
		return
	}
	h.state = nextFlow
}
