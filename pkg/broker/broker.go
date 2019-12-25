package broker

import (
	"fmt"
	"github.com/squ94wk/mqtt-broker/pkg/connection"
	"net"

	"github.com/squ94wk/mqtt-broker/pkg/actor"
	"github.com/squ94wk/mqtt-broker/pkg/client"
	"github.com/squ94wk/mqtt-broker/pkg/listener"
	"github.com/squ94wk/mqtt-broker/pkg/log"
	"github.com/squ94wk/mqtt-broker/pkg/session"
	"go.uber.org/zap"
)

type Broker struct {
	log *zap.Logger
}

func NewBroker() Broker {
	return Broker{
		log: log.Logger,
	}
}

func (r Broker) Start() {
	listenerHandlers := actor.NewScaleGroup(func() actor.Actor {
		handler := listener.NewHandler(&r, r.log)
		return &handler
	})

	//client := actor.NewScaleGroup(func() actor.Actor {
	//actor := action.NewHandler(&r, r.log)
	//return &actor
	//})

	sessionHandlers := actor.NewScaleGroup(func() actor.Actor {
		handler := session.NewHandler(&r, r.log)
		return &handler
	})

	listenerHandlers.Add()
	//client.Add()
	sessionHandlers.Add()

	select {}
}

func (r *Broker) Error(err error) {
	r.log.Error(fmt.Sprintf("error occured %v", err))
}

func (r *Broker) OnNewConnection(conn net.Conn) {
	r.log.Info("new connection")
	newConn := connection.NewConnection(r, conn, r.log)
	handler := client.NewHandler(r, newConn, r.log)

	go newConn.Start()
	go handler.Start()
}

func (r *Broker) SessionTakeover(s string) {

}
