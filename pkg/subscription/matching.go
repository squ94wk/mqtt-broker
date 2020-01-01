package subscription

//
//import "github.com/squ94wk/mqtt-common/pkg/topic"
//
//type Matcher struct {
//	direct       []Subscription
//	wildcard     []Matcher
//	deepWildcard []Subscription
//}
//
//func (m Matcher) Match(topic topic.Topic, l int) []Subscription {
//	if l == len(topic.Levels())-1 {
//		return m.direct
//	}
//
//	level := topic.Levels()[l]
//	if level == "#" {
//		var subs []Subscription
//		for _, matcher := range m.children {
//			subs = append(subs, matcher.Match(topic, l+1)...)
//		}
//		return subs
//	}
//
//	if level == "*" {
//		return m.AllSubsRecursive()
//	}
//
//	matcher, ok := m.children[level]
//	if ok {
//		return matcher.Match(topic, l+1)
//	}
//
//	return []Subscription{}
//}
//
//func (m Matcher) AllSubsRecursive() []Subscription {
//	subs := m.direct
//	for _, matcher := range m.children {
//		subs = append(subs, matcher.AllSubsRecursive()...)
//	}
//	return subs
//}
//
//func (m Matcher) SeekMatcher(filter topic.Filter) *Matcher {
//
//}
