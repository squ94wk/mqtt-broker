package session

import (
	"github.com/squ94wk/mqtt-broker/pkg/message"
	"sync"
)

type Session struct {
	clientId string
	sync.Mutex
	client client
}

type client interface {
	Deliver(message.Message)
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
