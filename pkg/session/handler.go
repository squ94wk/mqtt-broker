package session

import (
	"go.uber.org/zap"
)

type parent interface {
	Error(error)
	SessionTakeover(clientId string)
}

type Handler struct {
	parent   parent
	shutdown chan struct{}
	log      *zap.Logger
}

var (
	store   map[string]*Session
	actions chan func(h *Handler) error
)

func init() {
	store = make(map[string]*Session)
	actions = make(chan func(*Handler) error, 16)
}

func NewHandler(parent parent, log *zap.Logger) Handler {
	return Handler{
		parent:   parent,
		shutdown: make(chan struct{}, 1),
		log:      log,
	}
}

func (h *Handler) Start() {
	for {
		select {
		case _ = <-h.shutdown:
			return

		case action := <-actions:
			err := action(h)
			if err != nil {
				h.parent.Error(err)
			}
		}
	}
}

func (h Handler) Shutdown() {
	h.shutdown <- struct{}{}
}

func BindToSession(clientId string, cleanStart bool, client client, callback func(*Session, bool, error)) {
	actions <- func(h *Handler) error {
		var session *Session
		presentSession, ok := store[clientId]

		switch {
		case ok && cleanStart: // replace with clean session
			h.log.Debug("found session, but want clean start")
			CleanSession(presentSession)

			session = NewLocalSession(clientId, client)
			store[clientId] = session
			break

		case ok && !cleanStart: // take over session
			h.log.Debug("found session, want takeover")
			oldClient := session.client
			session.Replace(client)
			SessionTakeover(clientId, oldClient)
			break

		case !ok: // put clean session
			h.log.Debug("didn't find session, start clean", zap.String("clientId", clientId))
			session = NewLocalSession(clientId, client)
			store[clientId] = session
		}

		callback(session, ok && !cleanStart, nil)
		return nil
	}
}

func CleanSession(s *Session) {
	actions <- func(h *Handler) error {
		h.log.Debug("clean session")
		// deliver will message
		err := s.client.Close()
		// remove subscriptions
		return err
	}
}

func SessionTakeover(clientId string, client client) {
	actions <- func(h *Handler) error {
		// deliver will message
		// happily close conn
		return nil
	}
}
