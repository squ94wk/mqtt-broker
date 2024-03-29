package session

import (
	"time"

	"github.com/squ94wk/mqtt-broker/pkg/client"
	"github.com/squ94wk/mqtt-broker/pkg/subscription"
)

type Session struct {
	clientID      string // existence of a session is only based on this
	subscriptions map[string]subscription.Subscription
	lastActive    time.Time // for keep alive monitoring
	client        *client.Client
}
