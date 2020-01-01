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

func NewSubscription(filter topic.Filter, maxQoS byte, noLocal bool, retainAsPublished bool) Subscription {
	return Subscription{
		filter:            filter,
		maxQoS:            maxQoS,
		noLocal:           noLocal,
		retainAsPublished: retainAsPublished,
	}
}
