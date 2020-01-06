package client

import (
	"fmt"

	"github.com/squ94wk/mqtt-broker/pkg/message"
	"github.com/squ94wk/mqtt-broker/pkg/subscription"
	"github.com/squ94wk/mqtt-common/pkg/packet"
	"github.com/squ94wk/mqtt-common/pkg/topic"
	"go.uber.org/zap"
)

type connected struct{}

func (c connected) onPacket(client *Client, pkt packet.Packet) (state, error) {
	switch pkt.(type) {
	case *packet.Subscribe:
		subscribe := pkt.(*packet.Subscribe)
		reasons := make([]packet.SubackReason, len(subscribe.SubscriptionFilters()))
		for i, filter := range subscribe.SubscriptionFilters() {
			parsedFilter, err := topic.ParseFilter(filter.Filter())
			if err != nil {
				reasons[i] = packet.SubackTopicFilterInvalid
				continue
			}
			sub := subscription.NewSubscription(parsedFilter, filter.MaxQoS(), filter.NoLocal(), filter.RetainAsPublished())
			replacedSub, grantedQoS, err := client.parent.PerformSubscribe(client.clientID, sub, client)
			switch {
			case filter.RetainHandling() == packet.RetainHandlingAlways:
				fallthrough
			case filter.RetainHandling() == packet.RetainHandlingIfNotPresent && replacedSub:
				client.log.Debug("todo: queue retained message of topics that match filter", zap.String("filter", filter.Filter()))
				//sendRetainedMessage()
			}
			if err != nil {
				reasons[i] = packet.SubackImplementationSpecificError
				continue
			}

			switch grantedQoS {
			case 0:
				reasons[i] = packet.SubackGrantedQoS0

			case 1:
				reasons[i] = packet.SubackQoS1Granted

			case 2:
				reasons[i] = packet.SubackGrantedQoS2

			default:
				panic(fmt.Sprintf("invalid value %d for QoS granted", grantedQoS))
			}
		}
		var suback packet.Suback
		suback.SetPacketID(subscribe.PacketID())
		suback.SetReasons(reasons)
		_, err := suback.WriteTo(client.conn)
		if err != nil {
			return nil, err
		}
		return c, nil

	//case *packet.Unsubscribe:
	case *packet.Publish:
		publish := pkt.(*packet.Publish)
		msg := message.NewMessage(client.clientID, publish.Topic(), publish.QoS(), publish.Payload())
		if publish.Retain() {
			//client.parent.RetainMessage(msg)
		}
		client.parent.InboundMessage(msg)
		return c, nil
	//case *packet.PubAck:
	//case *packet.PubRec:
	//case *packet.PubComp:
	//case *packet.Ping:
	//case *packet.Pong:
	//case *packet.Disconnect:
	default:
		return nil, fmt.Errorf("TODO: expect Subscribe, Unsubscribe, Publish, Puback, Pubrec, Pubcomp, Ping, Pong or Disconnect")
	}
	return c, nil
}

func (c connected) onError(client *Client, err error) state {
	client.parent.Error(err)
	return c
}
