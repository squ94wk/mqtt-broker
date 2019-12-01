package action

import (
	"testing"
	"time"

	"github.com/squ94wk/mqtt-common/pkg/packet"
	"go.uber.org/zap"
)

var log *zap.Logger

func init() {
	log, _ = zap.NewDevelopment()
}

func Test(t *testing.T) {
	type args struct {
		client client
	}

	t.Run("connect must be first packet", func(t *testing.T) {
		c := newTestClient(func(in chan packet.Packet, err chan error) {
			var connack packet.Connack
			in <- connack
		})
		handler := NewHandler(testParent(func(err error) { t.Error(err) }), c, log)
		go handler.Start()

		time.Sleep(time.Second)
		handler.Shutdown()
		if !c.Closed() {
			t.Error()
		}
	})
}

type testParent func(error)

func (p testParent) Error(err error) {
	p(err)
}

type testClient struct {
	in     chan packet.Packet
	err    chan error
	out    chan packet.Packet
	closed bool
}

func newTestClient(init func(in chan packet.Packet, err chan error)) testClient {
	c := testClient{
		in:  make(chan packet.Packet, 0),
		err: make(chan error, 0),
		out: make(chan packet.Packet, 32),
	}
	go init(c.in, c.err)
	return c
}

func (c testClient) Packets() <-chan packet.Packet {
	return c.in
}

func (c testClient) Errors() <-chan error {
	return c.err
}

func (c testClient) Deliver(pkt packet.Packet) {
	c.out <- pkt
}

func (c testClient) Close() error {
	c.closed = true
	return nil
}

func (c testClient) Closed() bool {
	return c.closed
}
