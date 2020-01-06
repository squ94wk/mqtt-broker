package broker

import (
	"fmt"
	"net"

	"github.com/squ94wk/mqtt-broker/pkg/client"
	"github.com/squ94wk/mqtt-broker/pkg/delivery"
	"github.com/squ94wk/mqtt-broker/pkg/listener"
	"github.com/squ94wk/mqtt-broker/pkg/log"
	"github.com/squ94wk/mqtt-broker/pkg/message"
	"github.com/squ94wk/mqtt-broker/pkg/session"
	"github.com/squ94wk/mqtt-broker/pkg/subscription"
	"github.com/squ94wk/mqtt-common/pkg/packet"
	"go.uber.org/zap"
)

type Broker struct {
	sessionStore *session.Store
	listener     *listener.Handler
	delivery     *delivery.Handler
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
	b.delivery = delivery.NewHandler(b, b.log)

	go b.listener.Start()
	go b.sessionStore.Start()
	go b.delivery.Start()

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

func (b *Broker) PerformSubscribe(clientID string, sub subscription.Subscription, client *client.Client) (bool, packet.SubackReason, error) {
	replaced, err := b.sessionStore.RegisterSubscription(sub, clientID)
	b.delivery.AddSubscription(sub, clientID, client)
	if err != nil {
		return replaced, packet.SubackImplementationSpecificError, nil
	}
	return replaced, packet.SubackGrantedQoS0, nil
}

func (b *Broker) InboundMessage(msg message.Message) {
	b.delivery.Deliver(msg)
}

func (b *Broker) cleanUpSession(clientID string) {

}

func (b *Broker) replaceClient(clientID string, client *client.Client) {

}
