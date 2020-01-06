package delivery

import (
	"sync"

	"github.com/squ94wk/mqtt-broker/pkg/message"
	"github.com/squ94wk/mqtt-broker/pkg/subscription"
)

type Matcher struct {
	depth    int
	matches  map[string]Match
	children map[string]*Matcher
	sync.RWMutex
}

type Deliverer interface {
	Deliver(message.Message)
}

type Match struct {
	deliverer Deliverer
	sub       subscription.Subscription
}

func (m *Matcher) FindMatches(levels []string, matches map[Deliverer][]subscription.Subscription) {
	m.RLock()
	defer func() {
		m.RUnlock()
	}()
	if m.depth == len(levels) {
		for _, match := range m.matches {
			matches[match.deliverer] = append(matches[match.deliverer], match.sub)
		}
	}

	hashMatcher, ok := m.children["#"]
	if ok {
		for _, match := range hashMatcher.matches {
			matches[match.deliverer] = append(matches[match.deliverer], match.sub)
		}
	}
	plusMatcher, ok := m.children["+"]
	if ok {
		plusMatcher.FindMatches(levels, matches)
	}

	level := levels[m.depth]
	matcher, ok := m.children[level]
	if ok {
		matcher.FindMatches(levels, matches)
	}
}

func (m *Matcher) traverse(filter []string, createNonExisting bool) *Matcher {
	if len(filter) == 0 {
		return m
	}

	level := filter[0]
	matcher, ok := m.children[level]
	if !ok {
		if !createNonExisting {
			return nil
		}
		matcher = newMatcher(m.depth + 1)
		m.Lock()
		m.children[level] = matcher
		m.Unlock()
	}
	return matcher.traverse(filter[1:], createNonExisting)
}

func (m *Matcher) addSubscription(sub subscription.Subscription, clientID string, deliverer Deliverer) {
	levels := sub.Filter().Levels()
	matcher := m.traverse(levels, true)
	match := Match{
		deliverer: deliverer,
		sub:       sub,
	}
	m.Lock()
	matcher.matches[clientID] = match
	m.Unlock()
	//filter := sub.Filter()
	//if depth == len(filter.Levels()) {
	//	m.setMatch(clientID, FindMatches{
	//		deliverer: deliverer,
	//		sub:       sub,
	//	})
	//	return
	//}
	//
	//level := filter.Levels()[depth]
	//switch level {
	//case "#":
	//	m.setWildcardMatch(clientID, FindMatches{
	//		deliverer: deliverer,
	//		sub:       sub,
	//	})
	//
	//case "+":
	//	if m.singleWildcardMatcher == nil {
	//		m.Lock()
	//		m.singleWildcardMatcher = newMatcher()
	//		m.Unlock()
	//	}
	//	m.singleWildcardMatcher.addSubscription(sub, clientID, deliverer, depth+1)
	//
	//default:
	//	matcher, ok := m.children[level]
	//	if !ok {
	//		m.Lock()
	//		matcher = newMatcher()
	//		m.Unlock()
	//	}
	//	matcher.addSubscription(sub, clientID, deliverer, depth+1)
	//}
}

//func (m *Matcher) Replace
//
//func (m *Matcher) setMatch(clientID string, match FindMatches) {
//	m.Lock()
//	m.matches[clientID] = match
//	m.Unlock()
//}
//
//func (m *Matcher) setWildcardMatch(clientID string, match FindMatches) {
//	m.Lock()
//	m.wildcardMatches[clientID] = match
//	m.Unlock()
//}
//
//func (m *Matcher) RemoveMatch(clientID string) {
//	m.Lock()
//	delete(m.matches, clientID)
//	m.Unlock()
//}
//
//func (m *Matcher) RemoveWildcardMatch(clientID string) {
//	m.Lock()
//	delete(m.wildcardMatches, clientID)
//	m.Unlock()
//}

func newMatcher(depth int) *Matcher {
	return &Matcher{
		depth:    depth,
		matches:  make(map[string]Match),
		children: make(map[string]*Matcher),
	}
}

func (m Match) Deliverer() Deliverer {
	return m.deliverer
}

func (m Match) Sub() subscription.Subscription {
	return m.sub
}
