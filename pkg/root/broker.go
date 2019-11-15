package broker

import (
	"net"

	"github.com/squ94wk/mqtt-broker/pkg/client"
	"go.uber.org/zap"
)

type Broker struct {
	config BrokerConfig
	log    zap.Logger
}

type BrokerConfig struct {
	address string
}

func (r Broker) Error(err error) {
	r.log.Error(err)
}

func (r Broker) NewConnection(conn net.Conn) {
	connHandler := client.ConnHandler{conn: conn, parent: r}
	connHandler.Start()
}

func (r Broker) Start() {
	listener := client.Listener{parent: r, address: r.config.address}
}
