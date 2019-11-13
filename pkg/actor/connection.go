package actor

import (
	"fmt"
	"io"
	"net"

	"github.com/squ94wk/mqtt-common/pkg/packet"
)

type ConnectionParent interface {
	Error(error)
	OnPacket(packet.Packet, string)
}

type Connection struct {
	conn   net.Conn
	parent ConnectionParent
}

func (c *Connection) Start() {
	go c.write()
	go c.read()
}

func (c *Connection) read() {
	func() {
		for {
			pkt, err := c.readNext()
			if err != nil {
				c.parent.Error(fmt.Errorf("failed to read packet from conn: %v", err))
				break
			}

			//c.parent.OnPacket(pkt, c.out)
		}
	}()
}

func (c *Connection) write() {
	func() {
		for {
			var pkt packet.Packet
			select {
			case pkt = <-c.out.queue:
				err := pkt.Write(c.conn)
				if err != nil {
					c.parent.Error(fmt.Errorf("failed to write packet to conn: %v", err))
					break
				}
			}
		}
	}()
}

func (c *Connection) readNext() (packet.Packet, error) {
	var header packet.Header
	if err := packet.ReadHeader(c.conn, &header); err != nil {
		return nil, fmt.Errorf("failed to read packet: failed to read header: %v", err)
	}

	pkt, err := readRestOfPacket(c.conn, header)
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
		log.Info("read Connect packet")
		return &connect, nil

	case packet.CONNACK:
		panic("implement me")

	case packet.PUBLISH:
		//var publish packet.Publish
		//err := packet.ReadPublish(reader, &publish, header)
		//if err != nil {
		//return nil, fmt.Errorf("failed to read Publish packet: %v", err)
		//}
		//log.Info("read Publish packet")
		//return &publish, nil

	case packet.PUBACK:
		panic("implement me")

	case packet.PUBREC:
		panic("implement me")

	case packet.PUBREL:
		panic("implement me")

	case packet.PUBCOMP:
		panic("implement me")

	case packet.SUBSCRIBE:
		panic("implement me")

	case packet.SUBACK:
		panic("implement me")

	case packet.UNSUBSCRIBE:
		panic("implement me")

	case packet.UNSUBACK:
		panic("implement me")

	case packet.PINGREQ:
		panic("implement me")

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

	default:
		return nil, fmt.Errorf("header with invalid packet type '%v'", header.MsgType())
	}
}
