package connection

import (
	"fmt"
	"io"
	"net"

	"github.com/squ94wk/mqtt-common/pkg/packet"
	"go.uber.org/zap"
)

type parent interface {
	Error(error)
}

type Connection struct {
	parent      parent
	conn        net.Conn
	buf         chan func() (packet.Packet, error)
	actions     chan func() error
	finalAction func() error
	shutdown    chan struct{}
	log         *zap.Logger
}

func NewConnection(parent parent, conn net.Conn, log *zap.Logger) Connection {
	return Connection{
		parent:   parent,
		conn:     conn,
		buf:      make(chan func() (packet.Packet, error), 4),
		actions:  make(chan func() error, 4),
		shutdown: make(chan struct{}, 1),
		log:      log,
	}
}

func (c *Connection) Start() {
	defer func() {
		recover()
		err := c.conn.Close()
		if err != nil {
			c.parent.Error(fmt.Errorf("failed to Close() client: %v", err))
		}
		c.log.Debug("disconnected client")
	}()

	c.log.Debug("client started")
	wait := make(chan struct{}, 1)
	go func() {
		defer func() {
			recover()
		}()
		c.read()
		wait <- struct{}{}
	}()

	for action := range c.actions {
		err := action()
		if err != nil {
			c.handleError(err)
			break
		}
	}

	if c.finalAction != nil {
		err := c.finalAction()
		if err != nil {
			c.parent.Error(fmt.Errorf("failed to perform final action: %v", err))
		}
	}
	<-wait
}

func (c Connection) Packets() <-chan func() (packet.Packet, error) {
	return c.buf
}

func (c Connection) Deliver(pkt packet.Packet) {
	c.actions <- func() error {
		c.log.Debug("write packet")
		_, err := pkt.WriteTo(c.conn)
		return err
	}
}

//DeliverFinally instructs that a packet is delivered after all actions are performed
//and just before the tcp connection is closed.
func (c *Connection) DeliverFinally(pkt packet.Packet) {
	c.finalAction = func() error {
		c.log.Debug("final action: deliver packet")
		_, err := pkt.WriteTo(c.conn)
		return err
	}
}

func (c Connection) Close() error {
	c.log.Debug("closing client")
	close(c.actions)
	return nil
}

func (c Connection) read() {
	for {
		pkt, err := packet.ReadPacket(c.conn)
		c.buf <- func() (packet.Packet, error) {
			return pkt, err
		}

		if err != nil {
			return
		}
		c.log.Debug("read packet", zap.String("packet", fmt.Sprintf("%+v", pkt)))
	}
}

func (c Connection) handleError(err error) {
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

	//err = c.conn.Close()
	//if err != nil {
	//	c.parent.Error(fmt.Errorf("failed to close connection after error: %v", err))
	//}
}
