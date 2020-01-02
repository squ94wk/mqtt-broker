package client

import (
	"fmt"

	"github.com/squ94wk/mqtt-common/pkg/packet"
	"github.com/squ94wk/mqtt-common/pkg/topic"
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
			grantedQoS, err := client.parent.PerformSubscribe(client.clientID, subscribe.PacketID(), parsedFilter, filter.MaxQoS(), filter.NoLocal(), filter.RetainAsPublished(), filter.RetainHandling(), client)
			//switch {
			//case retainHandling == packet.RetainHandlingAlways: fallthrough
			//case retainHandling == packet.RetainHandlingIfNotPresent:
			//	sendRetainedMessage()
			//}
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
	//case *packet.Publish:
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
