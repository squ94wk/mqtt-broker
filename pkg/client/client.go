package client

import (
	"fmt"
	"io"
	"net"

	"github.com/squ94wk/mqtt-common/pkg/packet"
)

type ClientParent interface {
	Error(error)
	OnPacket(packet.Packet)
}

type Client struct {
	parent     ClientParent
	conn       net.Conn
	shutdown   chan bool
	packetsOut chan packet.Packet
}

func NewClient(conn net.Conn, parent ClientParent) Client {
	return Client{
		parent:   parent,
		conn:     conn,
		shutdown: make(chan bool, 1),
	}
}

func (h Client) Start() {
	go h.write()
	h.read()
}

func (h Client) read() {
	for {
		select {
		case _ = <-h.shutdown:
			break

		default:
			pkt, err := h.readNext()
			if err != nil {
				h.parent.Error(fmt.Errorf("failed to read packet from conn: %v", err))

				err := h.conn.Close()
				if err != nil {
					h.parent.Error(fmt.Errorf("failed to close connection after error: %v", err))
				}
				break
			}

			h.parent.OnPacket(pkt)
		}
	}
}

func (h Client) write() {
	for {
		var pkt packet.Packet
		select {
		case pkt = <-h.packetsOut:
			err := pkt.Write(h.conn)
			if err != nil {
				h.parent.Error(fmt.Errorf("failed to write packet to conn: %v", err))
				break
			}
		}
	}
}

func (h Client) readNext() (packet.Packet, error) {
	var header packet.Header
	if err := packet.ReadHeader(h.conn, &header); err != nil {
		return nil, fmt.Errorf("failed to read packet: failed to read header: %v", err)
	}

	pkt, err := readRestOfPacket(h.conn, header)
	if err != nil {
		return nil, err
	}

	return pkt, nil
}

func readRestOfPacket(reader io.Reader, header packet.Header) (packet.Packet, error) {
	switch header.MsgType() {
	case packet.CONNECT:
		if header.Flags() != 0 {
			return nil, fmt.Errorf("failed to read packet: invalid fixed header of Connect packet: invalid flags '%d'", header.Flags())
		}
		var connect packet.Connect
		err := packet.ReadConnect(reader, &connect, header)
		if err != nil {
			return nil, fmt.Errorf("failed to read Connect packet: %v", err)
		}
		return &connect, nil

	case packet.PUBLISH:
		//var publish packet.Publish
		//err := packet.ReadPublish(reader, &publish, header)
		//if err != nil {
		//return nil, fmt.Errorf("failed to read Publish packet: %v", err)
		//}
		//log.Info("read Publish packet")
		//return &publish, nil

	case packet.CONNACK:
	case packet.PUBACK:
	case packet.PUBREC:
	case packet.PUBREL:
	case packet.PUBCOMP:
	case packet.SUBSCRIBE:
	case packet.SUBACK:
	case packet.UNSUBSCRIBE:
	case packet.UNSUBACK:
	case packet.PINGREQ:
	case packet.PINGRESP:
		panic("implement me")

	case packet.DISCONNECT:
		//if header.Flags() != 0 {
		//return nil, fmt.Errorf("failed to read packet: invalid fixed header of Disconnect packet: invalid flags '%d'", header.Flags())
		//}
		//var disconnect packet.Disconnect
		//err := packet.ReadDisconnect(reader, &disconnect, header)
		//if err != nil {
		//return nil, fmt.Errorf("failed to read Disconnect packet: %v", err)
		//}
		//log.Info("read Disconnect packet")
		//return &disconnect, nil

	case packet.AUTH:
		panic("implement me")

	}
	return nil, fmt.Errorf("header with invalid packet type '%v'", header.MsgType())
}
