package client

import (
	"fmt"
	"net"

	"github.com/squ94wk/mqtt-common/pkg/packet"
	"go.uber.org/zap"
)

type ClientParent interface {
	Error(error)
	OnPacket(packet.Packet)
}

type Client struct {
	parent     ClientParent
	conn       net.Conn
	shutdown   chan bool
	log        *zap.Logger
	packetsOut chan packet.Packet
}

func NewClient(conn net.Conn, parent ClientParent, log *zap.Logger) Client {
	return Client{
		parent:     parent,
		conn:       conn,
		shutdown:   make(chan bool, 1),
		log:        log,
		packetsOut: make(chan packet.Packet, 4),
	}
}

func (h Client) Start() {
	h.log.Debug("client started")
	go h.write()
	h.read()
}

func (h Client) read() {
	for {
		select {
		case _ = <-h.shutdown:
			break

		default:
			pkt, err := packet.ReadPacket(h.conn)
			if err != nil {
				h.parent.Error(fmt.Errorf("failed to read packet from conn: %v", err))
				err := h.conn.Close()
				if err != nil {
					h.parent.Error(fmt.Errorf("failed to close connection after error: %v", err))
				}
				return
			}

			h.log.Debug("read packet")
			h.parent.OnPacket(pkt)
		}
	}
}

func (h Client) write() {
	for {
		var pkt packet.Packet
		select {
		case pkt = <-h.packetsOut:
			h.log.Debug("write packet")
			err := pkt.Write(h.conn)
			if err != nil {
				h.parent.Error(fmt.Errorf("failed to write packet to conn: %v", err))
				err := h.conn.Close()
				if err != nil {
					h.parent.Error(fmt.Errorf("failed to close connection after error: %v", err))
				}
				return
			}
		}
	}
}
