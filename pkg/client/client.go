package client

import (
	"fmt"
	"io"
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
	buf      chan func() (packet.Packet, error)
	actions  chan func() error
	shutdown chan struct{}
	log      *zap.Logger
}

func NewClient(parent clientParent, conn net.Conn, log *zap.Logger) Client {
	return Client{
		parent:   parent,
		conn:     conn,
		buf:      make(chan func() (packet.Packet, error), 4),
		actions:  make(chan func() error, 4),
		shutdown: make(chan struct{}, 1),
		log:      log,
	}
}

func (c Client) Start() {
	c.log.Debug("client started")
	wait := make(chan struct{}, 1)
	go func() {
		defer func() {
			recover()
			wait <- struct{}{}
		}()
		c.read()
	}()

	for action := range c.actions {
		err := action()
		if err != nil {
			c.handleError(err)
			c.shutdown <- struct{}{}
			break
		}
	}

	<-wait
	err := c.conn.Close()
	if err != nil {
		c.parent.Error(fmt.Errorf("failed to Close() client: %v", err))
	}
}

func (c Client) Packets() <-chan func() (packet.Packet, error) {
	return c.buf
}

func (c Client) Deliver(pkt packet.Packet) {
	c.actions <- func() error {
		c.log.Debug("write packet")
		return pkt.Write(c.conn)
	}
}

func (c Client) Close() error {
	c.log.Debug("closing client")
	close(c.actions)

	return nil
}

func (c Client) read() {
	for {
		select {
		case _ = <-c.shutdown:
			return

		default:
			pkt, err := packet.ReadPacket(c.conn)
			c.buf <- func() (packet.Packet, error) {
				return pkt, err
			}

			if err != nil {
				return
			}
			c.log.Debug("read packet", zap.Stringer("packet", pkt))
		}
	}
}

func (c Client) handleError(err error) {
	switch {
	case err == io.ErrUnexpectedEOF:
		c.log.Info("unexpected EOF")
		break
	case err == io.EOF:
		c.log.Info("EOF")
		break
	default:
		c.log.Warn("unexpected error occured", zap.Error(err))
	}

	err = c.conn.Close()
	if err != nil {
		c.parent.Error(fmt.Errorf("failed to close connection after error: %v", err))
	}
}
