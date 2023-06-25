package eventbus

import (
	"context"
	"fmt"
	"sync"

	"github.com/almahoozi/go-eventbus/pkg/id"
	"golang.org/x/sync/semaphore"
)

type bus struct {
	observers       map[string]observerWithOptions
	subscriptions   map[Stringer][]*subscription
	wg              sync.WaitGroup
	close           chan struct{}
	concurrency     int64
	continueOnError bool
}

func New(opts ...busOpt) *bus {
	b := &bus{
		observers:     make(map[string]observerWithOptions),
		subscriptions: make(map[Stringer][]*subscription),
		close:         make(chan struct{}),
		concurrency:   10,
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

// Subscribes to an event by name.
func (b *bus) On(name Stringer) *subscription {
	s := subscription{
		id:       id.New(),
		matchers: []Matcher{ExactMatcher(name)},
	}
	b.subscriptions[name] = append(b.subscriptions[name], &s)
	return &s
}

// Subscribes to an event by arbitrary matchers.
func (b *bus) When(matchers ...Matcher) *subscription {
	s := subscription{
		id:       id.New(),
		matchers: matchers,
	}
	// We don't want to accidentally match on the string for non-string matchers.
	key := noMatch("id:" + s.id)
	b.subscriptions[key] = append(b.subscriptions[key], &s)
	return &s
}

// Publishes an event with the provided name and data.
func (b *bus) Publish(ctx context.Context, name Stringer, data interface{}, opts ...eventOpt) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if b.closed() {
		return ErrBusClosed
	}

	e := newEvent(name, data)
	for _, opt := range opts {
		opt(&e)
	}

	b.wg.Add(1)
	defer b.wg.Done()

	return doWithTimeout(ctx, e.publishTimeout, func(ctx context.Context) error {
		if err := b.publishToObservers(ctx, e); err != nil {
			return err
		}

		return b.publishToSubscriptions(ctx, e)
	})
}

func (b *bus) publishToObservers(ctx context.Context, e Event) error {
	s := semaphore.NewWeighted(b.concurrency)
	for _, o := range b.observers {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := s.Acquire(ctx, 1)
		if err != nil {
			return err
		}

		o := o
		go func() {
			defer s.Release(1)
			_ = doWithTimeout(ctx, shortestDuration(e.handlerTimeout, o.opts.timeout), func(ctx context.Context) error {
				o.Observe(ctx, e.Name, e.Data)
				return nil
			})
		}()
	}

	return nil
}

func (b *bus) publishToSubscriptions(ctx context.Context, e Event) error {
	var errs Errors
	for _, subs := range b.subscriptions {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		for _, s := range subs {
			if !s.Match(e.Name, e.Data) {
				continue
			}

			for _, fn := range s.funcs {
				err := doWithTimeout(ctx, e.handlerTimeout, func(ctx context.Context) error {
					return fn(ctx, e.Name, e.Data)
				})
				if err != nil {
					if b.continueOnError {
						errs = append(errs, fmt.Errorf("subscription error; subscription: %v, event: %v: %w", s, e, err))
						continue
					}
					return err
				}
			}
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

// Adds an observer. Observers are notified of all published events, and are
// executed in parallel.
func (b *bus) AddObserver(o observer, opts ...observerOpt) string {
	id := id.New()

	options := observerOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	b.observers[id] = observerWithOptions{
		observer: o,
		opts:     options,
	}

	return id
}

// Removes an observer.
func (b *bus) RemoveObserver(id string) bool {
	if _, ok := b.observers[id]; ok {
		delete(b.observers, id)
		return true
	}
	return false
}

// Waits for all published events to finish processing.
func (b *bus) Flush(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}
	done := make(chan struct{})
	defer close(done)
	go func() {
		b.wg.Wait()
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
	case <-done:
	}
}

// Waits for the bus to be closed and then flushes.
func (b *bus) Wait(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}
	done := make(chan struct{})
	defer close(done)
	go func() {
		<-b.close
		b.Flush(ctx)
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
	case <-done:
	}
}

// Signals the bus to close.
func (b *bus) Close() {
	if b.closed() {
		return
	}
	close(b.close)
}

func (b *bus) closed() bool {
	select {
	case <-b.close:
		return true
	default:
		return false
	}
}
