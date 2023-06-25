// eventtbus is a package for a simple event bus.
package eventbus

import "context"

var _default = New()

// SetDefault sets the default event bus.
func SetDefault(eb *bus) {
	_default = eb
}

// On subscribes to an event by name in the default event bus.
func On(name Stringer) *subscription {
	return _default.On(name)
}

// When subscribes to an event by arbitrary matchers in the default event bus.
func When(matchers ...Matcher) *subscription {
	return _default.When(matchers...)
}

// Publishes an event with the provided name and data.
func Publish(ctx context.Context, name Stringer, data interface{}, opts ...eventOpt) error {
	return _default.Publish(ctx, name, data, opts...)
}

// Adds an observer. Observers are notified of all published events, and are
// executed in parallel.
func AddObserver(o observer, opts ...observerOpt) string {
	return _default.AddObserver(o, opts...)
}

// Removes an observer.
func RemoveObserver(id string) bool {
	return _default.RemoveObserver(id)
}

// Waits for all published events to finish processing.
func Flush(ctx context.Context) {
	_default.Flush(ctx)
}

// Waits for the bus to be closed and then flushes.
func Wait(ctx context.Context) {
	_default.Wait(ctx)
}

// Signals the bus to close.
func Close() {
	_default.Close()
}
