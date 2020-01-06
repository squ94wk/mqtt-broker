package delivery

import (
	"github.com/squ94wk/mqtt-broker/pkg/message"
	"github.com/squ94wk/mqtt-broker/pkg/subscription"
	"go.uber.org/zap"
)

type parent interface {
	Error(error)
}

var (
	actions     = make(chan func() error)
	rootMatcher = newMatcher(0)
)

type Handler struct {
	parent   parent
	shutdown chan struct{}
	log      *zap.Logger
}

func NewHandler(parent parent, log *zap.Logger) *Handler {
	return &Handler{
		parent:   parent,
		shutdown: make(chan struct{}, 1),
		log:      log,
	}
}

func (h Handler) Start() {
	for {
		select {
		case _ = <-h.shutdown:
			return

		case action := <-actions:
			err := action()
			if err != nil {
				return
			}
		}
	}
}

func (h *Handler) Deliver(msg message.Message) {
	actions <- func() error {
		matches := make(map[Deliverer][]subscription.Subscription)
		rootMatcher.FindMatches(msg.Topic().Levels(), matches)
		for deliverer, subs := range matches {
			for _, sub := range subs {
				effectiveQoS := effectiveQoS(msg.QoS(), sub.MaxQoS())
				outboundMsg := message.NewMessage(msg.Sender(), msg.Topic(), effectiveQoS, msg.Payload())
				deliverer.Deliver(outboundMsg)
			}
		}
		return nil
	}
}

func (h *Handler) AddSubscription(sub subscription.Subscription, clientID string, deliverer Deliverer) {
	actions <- func() error {
		rootMatcher.addSubscription(sub, clientID, deliverer)
		return nil
	}
}

func effectiveQoS(a byte, b byte) byte {
	if a > b {
		return b
	} else {
		return a
	}
}
