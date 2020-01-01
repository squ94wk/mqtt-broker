package broker

import (
	"fmt"
	"net"

	"github.com/squ94wk/mqtt-common/pkg/packet"
	"github.com/squ94wk/mqtt-common/pkg/topic"

	"github.com/squ94wk/mqtt-broker/pkg/client"
	"github.com/squ94wk/mqtt-broker/pkg/listener"
	"github.com/squ94wk/mqtt-broker/pkg/log"
	"github.com/squ94wk/mqtt-broker/pkg/session"
	"go.uber.org/zap"
)

type Broker struct {
	sessionStore *session.Store
	listener     *listener.Handler
	log          *zap.Logger
}

func NewBroker() Broker {
	return Broker{
		log: log.Logger,
	}
}

func (b *Broker) Start() {
	b.listener = listener.NewHandler(b, b.log)
	b.sessionStore = session.NewStore(b, b.log)

	go b.listener.Start()
	go b.sessionStore.Start()

	select {}
}

func (b *Broker) Error(err error) {
	b.log.Error(fmt.Sprintf("error occured %v", err))
}

func (b *Broker) OnNewConnection(conn net.Conn) {
	b.log.Info("new connection")
	handler := client.NewClient(b, conn, b.log)

	go handler.Start()
}

func (b *Broker) PerformConnect(clientID string, cleanStart bool, client *client.Client) (string, bool, error) {
	assignedID, oldClient, err := b.sessionStore.RegisterNewSession(clientID, cleanStart, client)
	if err != nil {
		b.log.Error("failed to connect", zap.Error(err))
	}

	if oldClient == nil {
		return assignedID, false, nil
	}

	if cleanStart {
		b.cleanUpSession(clientID)
	} else {
		b.replaceClient(clientID, client)
	}
	oldClient.Disconnect(packet.DisconnectSessionTakenOver, "")
	return assignedID, true, nil
}

func (b *Broker) PerformSubscribe(packetID uint16, filter topic.Filter, maxQoS byte, noLocal bool, retainAsPublished bool, retainHandling byte) (packet.SubackReason, error) {
	return packet.SubackGrantedQoS0, nil
}

func (b *Broker) cleanUpSession(clientID string) {

}

func (b *Broker) replaceClient(clientID string, client *client.Client) {

}
