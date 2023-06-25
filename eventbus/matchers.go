package eventbus

import (
	"regexp"
	"strings"
)

type (
	// Matcher is an interface that matches events.
	Matcher interface {
		Match(Stringer, interface{}) bool
		String() string
	}
	regexMatcher struct {
		str   string
		regex *regexp.Regexp
	}
	// PredicateMatcher is a function that accepts an event name and data and returns true if the
	// event matches the predicate.
	PredicateMatcher func(Stringer, interface{}) bool
	// StringMatcher is a string that matches events by name, ignoring case and type.
	StringMatcher string
	noMatch       string
)

func (m noMatch) String() string {
	return string(m)
}

// WildcardMatcher is a string that utilizes the asterisk (*) as a wildcard character.
// It can match all events with "*", all events with a prefix "foo*", all events
// with a suffix "*bar", all events with a substring "foo*bar", or a combination
// of the above. A question mark (?) can be used to match a single character.
func WildcardMatcher(s string) regexMatcher {
	s = strings.ReplaceAll(s, "*", ".*")
	s = strings.ReplaceAll(s, "?", ".?")
	return regexMatcher{
		str:   s,
		regex: regexp.MustCompile(s),
	}
}

// RegexMatcher is a string that utilizes a regular expression to match events.
func RegexMatcher(s string) (regexMatcher, error) {
	r, err := regexp.Compile(s)
	if err != nil {
		return regexMatcher{}, err
	}
	return regexMatcher{
		str:   s,
		regex: r,
	}, nil
}

func (m regexMatcher) Match(name Stringer, data interface{}) bool {
	return m.regex.MatchString(name.String())
}

func (m regexMatcher) String() string {
	return m.str
}

func (m PredicateMatcher) Match(name Stringer, data interface{}) bool {
	return m(name, data)
}

func (m PredicateMatcher) String() string {
	return "predicate"
}

func (m StringMatcher) Match(name Stringer, data interface{}) bool {
	return strings.EqualFold(name.String(), name.String())
}

func (m StringMatcher) String() string {
	return string(m)
}

// ExactMatcher is matcher that matches events by equality.
func ExactMatcher(thisName Stringer) PredicateMatcher {
	return func(otherName Stringer, data interface{}) bool {
		return thisName == otherName
	}
}
