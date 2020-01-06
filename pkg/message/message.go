package message

import "github.com/squ94wk/mqtt-common/pkg/topic"

//Message defines an mqtt application message.
type Message struct {
	sender  string
	topic   topic.Topic
	qos     byte
	payload []byte
}

//NewMessage is the constructor for the Message type.
func NewMessage(sender string, topic topic.Topic, qos byte, payload []byte) Message {
	return Message{
		sender:  sender,
		topic:   topic,
		qos:     qos,
		payload: payload,
	}
}

func (m Message) Sender() string {
	return m.sender
}

func (m Message) Topic() topic.Topic {
	return m.topic
}

func (m Message) QoS() byte {
	return m.qos
}

func (m Message) Payload() []byte {
	return m.payload
}
