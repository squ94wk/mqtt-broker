package session

import (
	"sync"

	"github.com/squ94wk/mqtt-common/pkg/packet"
)

type Session struct {
	clientId string
	sync.Mutex
	client client
}

type client interface {
	Deliver(packet.Packet)
	Close() error
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
