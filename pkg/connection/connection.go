package connection

import (
	"fmt"
	"io"
	"net"

	"github.com/squ94wk/mqtt-common/pkg/packet"
	"go.uber.org/zap"
)

func (c *Connection) Start() {

	c.log.Debug("client started")
	wait := make(chan struct{}, 1)
	go func() {
		defer func() {
			recover()
		}()
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
