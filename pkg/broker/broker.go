package broker

import (
	"fmt"
	"net"

	"github.com/squ94wk/mqtt-broker/pkg/action"
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
	listener := actor.NewScaleGroup(func() actor.Actor {
		actor := listener.NewHandler(&r, r.log)
		return &actor
	})

	//client := actor.NewScaleGroup(func() actor.Actor {
	//actor := action.NewHandler(&r, r.log)
	//return &actor
	//})

	session := actor.NewScaleGroup(func() actor.Actor {
		actor := session.NewHandler(&r, r.log)
		return &actor
	})

	listener.Add()
	//client.Add()
	session.Add()

	select {}
}

func (r *Broker) Error(err error) {
	r.log.Error(fmt.Sprintf("error occured %v", err))
}

func (r *Broker) OnNewConnection(conn net.Conn) {
	r.log.Info("new connection")
	client := client.NewClient(r, conn, r.log)
	handler := action.NewHandler(r, client, r.log)

	go client.Start()
	go handler.Start()
}

func (r *Broker) SessionTakeover(s string) {

}
