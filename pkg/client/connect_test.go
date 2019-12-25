package client

import (
	"testing"
	"time"

	"github.com/squ94wk/mqtt-broker/pkg/session"
	"github.com/squ94wk/mqtt-common/pkg/packet"
	"github.com/squ94wk/mqtt-common/pkg/topic"
	"go.uber.org/zap"
)

var (
	log, _ = zap.NewDevelopment()
)

func TestConnect(t *testing.T) {
	packets := make(chan func() (packet.Packet, error), 4)
	packets <- func() (packet.Packet, error) {
		var pkt packet.Connect
		pkt.SetKeepAlive(1000)
		pkt.SetCleanStart(true)
		pkt.Set
		pkt := packet.NewConnect(
			true,
			uint16(1000),
			nil,
			"client",
			false,
			packet.Qos0,
			nil,
			topic.Topic(""),
			nil,
			"",
			nil,
		)
		return pkt, nil
	}
	client := NewTestClient(packets)

	parent := TestClientParent{}
	handler := connect.NewHandler(&parent, &client, log)
	go func() {
		handler := session.NewHandler(&parent, log)
		handler.Start()
	}()
	wait := make(chan struct{})
	go func() {
		handler.Start()
		wait <- struct{}{}
	}()

	select {
	case _ = <-wait:
		break
	case _ = <-time.After(2 * time.Second):
		break
	}

	pkt, ok := <-client.out
	if !ok {
		t.Error("no packet received")
		return
	}
	if !parent.completed {
		t.Error("not completed")
	}
	if _, ok := pkt.(*packet.Connack); !ok {
		t.Error("not a connack")
	}
}

type TestClient struct {
	packets chan func() (packet.Packet, error)
	out     chan packet.Packet
}

func NewTestClient(packets chan func() (packet.Packet, error)) TestClient {
	return TestClient{
		packets: packets,
		out:     make(chan packet.Packet, 4),
	}
}

func (c *TestClient) Packets() <-chan func() (packet.Packet, error) {
	return c.packets
}

func (c *TestClient) Deliver(pkt packet.Packet) {
	c.out <- pkt
	//c.out = append(c.out, pkt)
}

func (c *TestClient) Close() error {
	return nil
}

type TestClientParent struct {
	completed bool
}

func (p *TestClientParent) Error(error) {}

func (p *TestClientParent) ConnectComplete(client Client) {
	p.completed = true
}

func (p *TestClientParent) SessionTakeover(s string) {
}
