package session

import (
	"github.com/squ94wk/mqtt-broker/pkg/message"
	"github.com/squ94wk/mqtt-common/pkg/packet"

	"sync"
)

type Session struct {
	clientId string
	sync.Mutex
	client client
}

type client interface {
	Deliver(message.Message)
	Disconnect(reason packet.DisconnectReason, reasonMsg string)
}

func (s Session) ClientID() string {
	return s.clientId
}

func NewLocalSession(clientId string, client client) *Session {
	return &Session{
		clientId: clientId,
		client:   client,
	}
}

func (s *Session) Replace(client client) {
	s.Lock()
	s.client = client
	s.Unlock()
}
