package session

import (
	"math/rand"
	"time"

	"github.com/squ94wk/mqtt-broker/pkg/client"
	"go.uber.org/zap"
)

type parent interface {
	Error(error)
}

type Store struct {
	sessions map[string]Session
	shutdown chan struct{}
	parent   parent
	log      *zap.Logger
}

var (
	actions chan func(*Store) error
)

func init() {
	actions = make(chan func(*Store) error, 16)
}

func NewStore(parent parent, log *zap.Logger) *Store {
	return &Store{
		sessions: make(map[string]Session),
		parent:   parent,
		log:      log,
	}
}

func (s *Store) Start() {
	for {
		select {
		case _ = <-s.shutdown:
			return

		case action := <-actions:
			err := action(s)
			if err != nil {
				s.parent.Error(err)
			}
		}
	}
}

func (s Store) Shutdown() {
	s.shutdown <- struct{}{}
}

func (s Store) RegisterNewSession(clientID string, cleanStart bool, client *client.Client) (assignedID string, oldClient *client.Client, callerErr error) {
	wait := make(chan struct{}, 1)
	actions <- func(s *Store) error {
		defer func() {
			wait <- struct{}{}
		}()
		if clientID == "" {
			clientID = s.getFreeClientID()
		}
		assignedID = clientID
		existingSession, ok := s.sessions[clientID]

		var newSession Session
		if cleanStart {
			newSession = Session{
				clientID:   clientID,
				lastActive: time.Now(),
				client:     client,
			}
		} else {
			newSession = copySession(existingSession)
		}
		s.sessions[clientID] = newSession

		if ok {
			oldClient = existingSession.client
		}
		return nil
	}
	//TODO: select on cancel context here?
	<-wait
	return
}

func (s Store) getFreeClientID() string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	randStringRunes := func(n int) string {
		b := make([]rune, n)
		for i := range b {
			b[i] = letterRunes[rand.Intn(len(letterRunes))]
		}
		return string(b)
	}
	for {
		clientID := randStringRunes(32)
		if _, ok := s.sessions[clientID]; !ok {
			return clientID
		}
	}
}

func copySession(s Session) Session {
	return Session{
		clientID:      s.clientID,
		subscriptions: s.subscriptions,
		lastActive:    time.Now(),
	}
}
