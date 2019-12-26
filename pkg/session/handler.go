package session

import (
	"github.com/squ94wk/mqtt-common/pkg/packet"
	"go.uber.org/zap"
)

type parent interface {
	Error(error)
	SessionTakeover(clientID string)
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

func BindToSession(clientID string, cleanStart bool, client client) (session *Session, sessionPresent bool, callerErr error) {
	wait := make(chan struct{}, 1)
	actions <- func(h *Handler) error {
		existingSession, ok := store[clientID]

		switch {
		case ok && cleanStart: // replace with clean session
			h.log.Debug("found session, but want clean start")
			CleanSession(existingSession)

			session = NewLocalSession(clientID, client)
			store[clientID] = session
			break

		case ok && !cleanStart: // take over session
			h.log.Debug("found session, want takeover")
			oldClient := session.client
			session.Replace(client)
			SessionTakeover(clientID, oldClient)
			break

		case !ok: // put clean session
			h.log.Debug("didn't find session, start clean", zap.String("clientID", clientID))
			session = NewLocalSession(clientID, client)
			store[clientID] = session
		}

		sessionPresent = ok && !cleanStart
		wait <- struct{}{}

		return nil
	}
	//TODO: select on cancel context here?
	<-wait
	return
}

func CleanSession(s *Session) {
	actions <- func(h *Handler) error {
		h.log.Debug("clean session")
		// remove subscriptions
		// look up will message
		// deliver will message
		s.client.Disconnect(packet.DisconnectSessionTakenOver, "")
		return nil
	}
}

func SessionTakeover(clientID string, client client) {
	actions <- func(h *Handler) error {
		// look up will message
		// deliver will message
		// disconnect gracefully
		client.Disconnect(packet.DisconnectSessionTakenOver, "")
		return nil
	}
}
