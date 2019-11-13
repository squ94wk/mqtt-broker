package actor

import "net"

type Root struct {
	config RootConfig
}

type RootConfig struct {
	address string
}

func (r Root) Error(err error) {
	log.Error(err)
}

func (r Root) NewConnection(conn net.Conn) {
	connection := ConnHandler{conn: conn, parent: r}
	connection.Start()
}

func (r Root) Start() {
	listener := Listener{parent: r, address: r.config.address}
}
