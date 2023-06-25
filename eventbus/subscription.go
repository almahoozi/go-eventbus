package eventbus

import "context"

type subscription struct {
	id       string
	matchers []Matcher
	funcs    []func(context.Context, Stringer, interface{}) error
}

// Or returns a new subscription that is the logical OR of the provided
// matchers.
func (s *subscription) Or(matcher Matcher) *subscription {
	s.matchers = append(s.matchers, matcher)
	return s
}

// Assigns the function to be executed when the event is published.
func (s *subscription) Do(fn func(context.Context, Stringer, interface{}) error) {
	s.funcs = append(s.funcs, fn)
}

// Match returns true if the event matches the subscription.
func (s *subscription) Match(name Stringer, data interface{}) bool {
	for _, m := range s.matchers {
		if m.Match(name, data) {
			return true
		}
	}
	return false
}

// String returns the subscription's ID.
func (s *subscription) String() string {
	return s.id
}
