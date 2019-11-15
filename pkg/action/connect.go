package action

import (
	"net"

	"github.com/squ94wk/mqtt-common/pkg/packet"
)

type Connect struct {
	packet           packet.Connect
	conn             net.Conn
	assignedClientId string
}

func NewConnect(c packet.Connect, conn net.Conn) {
	return Connect{packet: c, conn: conn, assignedClientId: ""}
}
