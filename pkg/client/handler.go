package client

import (
	"fmt"
	"net"

	"github.com/squ94wk/mqtt-broker/pkg/message"
	"github.com/squ94wk/mqtt-common/pkg/packet"
	"github.com/squ94wk/mqtt-common/pkg/topic"
	"go.uber.org/zap"
)

type parent interface {
	Error(error)
	PerformConnect(string, bool, *Client) (string, bool, error)
	PerformSubscribe(uint16, topic.Filter, byte, bool, bool, byte) (packet.SubackReason, error)
}

type Client struct {
	parent  parent
	conn    net.Conn
	state   state
	pkts    chan packet.Packet
	errs    chan error
	actions chan func() (bool, error)
	close   chan struct{}
	log     *zap.Logger
}

type state interface {
	onPacket(*Client, packet.Packet) (state, error)
	onError(*Client, error) state
}

func NewClient(parent parent, conn net.Conn, log *zap.Logger) *Client {
	return &Client{
		parent:  parent,
		conn:    conn,
		state:   initial{},
		actions: make(chan func() (bool, error), 4),
		close:   make(chan struct{}, 1),
		log:     log,
	}
}

func (c Client) Start() {
	c.log.Debug("client handler started")
	go func() {
		for {
			pkt, err := packet.ReadPacket(c.conn)
			if err != nil {
				c.log.Error(fmt.Sprintf("failed to read packet: %v", err))
				break
			}
			c.handleInboundPacket(pkt)
		}
		c.close <- struct{}{}
	}()

loop:
	for {
		select {
		case _ = <-c.close:
			break loop

		case action := <-c.actions:
			final, err := action()
			if err != nil {
				c.parent.Error(err)
				break loop
			}
			if final {
				break loop
			}
		}
	}

	err := c.conn.Close()
	if err != nil {
		c.parent.Error(fmt.Errorf("failed to close tcp connection: %v", err))
	}
	c.log.Debug("closed tcp connection")
}

func (c Client) Deliver(msg message.Message) {
	c.actions <- func() (bool, error) {
		//TODO: make publish packet
		//c.conn.Deliver(pkt)
		return false, nil
	}
}

func (c Client) Disconnect(reason packet.DisconnectReason, reasonMsg string) {
	c.actions <- func() (bool, error) {
		var disconnect packet.Disconnect
		disconnect.SetDisconnectReason(reason)
		props := packet.NewProperties()
		if reasonMsg != "" {
			props.Add(packet.NewProperty(packet.ReasonString, packet.StringPropPayload(reasonMsg)))
		}
		disconnect.SetProps(props)
		c.log.Debug("write packet")
		_, err := disconnect.WriteTo(c.conn)
		if err != nil {
			c.close <- struct{}{}
		}
		return true, nil
	}
}

func (c Client) connackWithError(err error) {
	var connack packet.Connack
	connack.SetSessionPresent(false)
	connack.SetConnectReason(packet.ConnectUnspecifiedError)
	connack.SetProps(packet.NewProperties(
		packet.NewProperty(
			packet.ReasonString,
			packet.StringPropPayload(fmt.Sprintf("%v", err)),
		),
	))
	_, err = connack.WriteTo(c.conn)
	if err != nil {
		c.close <- struct{}{}
	}
}

func (c Client) connackWithSuccess(clientID string, sessionPresent bool) {
	var connack packet.Connack
	connack.SetSessionPresent(sessionPresent)
	connack.SetProps(packet.NewProperties(
		packet.NewProperty(
			packet.AssignedClientIdentifier,
			packet.StringPropPayload(clientID),
		),
	))
	connack.SetConnectReason(packet.ConnectSuccess)
	_, err := connack.WriteTo(c.conn)
	if err != nil {
		c.close <- struct{}{}
	}
}

func (c *Client) handleInboundPacket(pkt packet.Packet) {
	nextState, err := c.state.onPacket(c, pkt)
	if err != nil {
		c.parent.Error(fmt.Errorf("failed to handle inbound packet: state reports error: %v", err))
		return
	}
	c.state = nextState
}
