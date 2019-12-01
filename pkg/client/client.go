package client

import (
	"fmt"
	"net"

	"github.com/squ94wk/mqtt-common/pkg/packet"
	"go.uber.org/zap"
)

type clientParent interface {
	Error(error)
}

type Client struct {
	parent   clientParent
	conn     net.Conn
	buf      chan packet.Packet
	err      chan error
	out      chan packet.Packet
	shutdown chan bool
	log      *zap.Logger
}

func NewClient(parent clientParent, conn net.Conn, log *zap.Logger) Client {
	return Client{
		parent:   parent,
		conn:     conn,
		buf:      make(chan packet.Packet, 4),
		out:      make(chan packet.Packet, 4),
		shutdown: make(chan bool, 1),
		log:      log,
	}
}

func (c Client) Start() {
	c.log.Debug("client started")
	go c.write()
	c.read()
}

func (c Client) Packets() <-chan packet.Packet {
	return c.buf
}

func (c Client) Errors() <-chan error {
	return c.err
}

func (c Client) Deliver(p packet.Packet) {
	c.out <- p
}

func (c Client) Close() error {
	c.log.Debug("closing client")
	err := c.conn.Close()
	if err != nil {
		return fmt.Errorf("failed to Close() client: %v", err)
	}
	return nil
}

func (c Client) read() {
	for {
		select {
		case _ = <-c.shutdown:
			break

		default:
			pkt, err := packet.ReadPacket(c.conn)
			if err != nil {
				c.parent.Error(fmt.Errorf("failed to read packet from conn: %v", err))
				err := c.conn.Close()
				if err != nil {
					c.parent.Error(fmt.Errorf("failed to close connection after error: %v", err))
				}
				return
			}

			c.log.Debug("read packet", zap.Stringer("packet", pkt))
			c.buf <- pkt
		}
	}
}

func (c Client) write() {
	for {
		var pkt packet.Packet
		select {
		case pkt = <-c.out:
			c.log.Debug("write packet")
			err := pkt.Write(c.conn)
			if err != nil {
				c.parent.Error(fmt.Errorf("failed to write packet to conn: %v", err))
				err := c.conn.Close()
				if err != nil {
					c.parent.Error(fmt.Errorf("failed to close connection after error: %v", err))
				}
				return
			}
		}
	}
}
