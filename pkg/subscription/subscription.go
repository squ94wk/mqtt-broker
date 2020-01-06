package subscription

import (
	"github.com/squ94wk/mqtt-common/pkg/topic"
)

type Subscription struct {
	filter            topic.Filter
	maxQoS            byte
	noLocal           bool
	retainAsPublished bool
	subscriptionID    uint32
}

func (s Subscription) Filter() topic.Filter {
	return s.filter
}

func (s Subscription) MaxQoS() byte {
	return s.maxQoS
}

func (s Subscription) NoLocal() bool {
	return s.noLocal
}

func (s Subscription) RetainAsPublished() bool {
	return s.retainAsPublished
}

func (s Subscription) SubscriptionID() uint32 {
	return s.subscriptionID
}

func NewSubscription(filter topic.Filter, maxQoS byte, noLocal bool, retainAsPublished bool) Subscription {
	return Subscription{
		filter:            filter,
		maxQoS:            maxQoS,
		noLocal:           noLocal,
		retainAsPublished: retainAsPublished,
	}
}
