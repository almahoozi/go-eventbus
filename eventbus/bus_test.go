package eventbus_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/almahoozi/go-eventbus/eventbus"
)

type (
	EventName       string
	ConstantMatcher struct {
		value bool
	}
)

func (e EventName) String() string {
	return string(e)
}

func (c ConstantMatcher) Match(eventbus.Stringer, interface{}) bool {
	return c.value
}

func (ConstantMatcher) String() string {
	return "ConstantMatcher"
}

var testEvent = EventName("test")

func TestOn_EventPublished_CallsDo(t *testing.T) {
	ctx := context.WithValue(context.Background(), testEvent, "test")
	called := false
	bus := eventbus.New()

	bus.On(testEvent).Do(func(ctx context.Context, name eventbus.Stringer, data interface{}) error {
		if ctx.Value(testEvent) != "test" {
			t.Error("expected correct context to be passed")
		}

		if name.String() != testEvent.String() {
			t.Error("expected correct event name to be passed")
		}

		if data != "some data" {
			t.Error("expected correct data to be passed")
		}
		called = true
		return nil
	})

	if err := bus.Publish(ctx, testEvent, "some data"); err != nil {
		t.Error("expected no error", err)
	}

	if !called {
		t.Error("expected event to be called")
	}
}

func TestOn_OtherEventPublished_DoesNotCallDo(t *testing.T) {
	ctx := context.Background()
	called := false
	bus := eventbus.New()
	s := bus.On(testEvent)
	s.Do(func(_ context.Context, _ eventbus.Stringer, _ interface{}) error {
		called = true
		return nil
	})

	if err := bus.Publish(ctx, EventName("other"), nil); err != nil {
		t.Error("expected no error", err)
	}

	if called {
		t.Error("expected event to not be called")
	}
}

func TestWhen_MatcherReturnsTrue_CallsDo(t *testing.T) {
	ctx := context.Background()
	called := false
	bus := eventbus.New()
	s := bus.When(ConstantMatcher{true})
	s.Do(func(_ context.Context, _ eventbus.Stringer, _ interface{}) error {
		called = true
		return nil
	})

	if err := bus.Publish(ctx, EventName("other"), nil); err != nil {
		t.Error("expected no error", err)
	}

	if !called {
		t.Error("expected event to be called")
	}
}

func TestWhen_MatcherReturnsFalse_DoesNotCallDo(t *testing.T) {
	ctx := context.Background()
	called := false
	bus := eventbus.New()
	s := bus.When(ConstantMatcher{false})
	s.Do(func(_ context.Context, _ eventbus.Stringer, _ interface{}) error {
		called = true
		return nil
	})

	if err := bus.Publish(ctx, EventName("other"), nil); err != nil {
		t.Error("expected no error", err)
	}

	if called {
		t.Error("expected event to not be called")
	}
}

func TestPublish_WithNoSubscribers_ReturnsNoError(t *testing.T) {
	ctx := context.Background()
	bus := eventbus.New()

	if err := bus.Publish(ctx, EventName("other"), nil); err != nil {
		t.Error("expected no error", err)
	}
}

func TestPublish_WithOneSubscriber_CallsDo(t *testing.T) {
	ctx := context.Background()
	bus := eventbus.New()
	called := false
	bus.On(testEvent).Do(func(_ context.Context, _ eventbus.Stringer, data interface{}) error {
		called = true
		if data != "data" {
			t.Error("expected correct data to be passed")
		}
		return nil
	})

	if err := bus.Publish(ctx, testEvent, "data"); err != nil {
		t.Error("expected no error", err)
	}

	bus.Flush(ctx)
	if !called {
		t.Error("expected Do to be called")
	}
}

func TestPublish_WithOneSubscriberThatReturnsError_ReturnsError(t *testing.T) {
	ctx := context.Background()
	bus := eventbus.New()
	bus.On(testEvent).Do(func(_ context.Context, _ eventbus.Stringer, _ interface{}) error {
		return errors.New("some error")
	})

	if err := bus.Publish(ctx, testEvent, nil); err == nil || err.Error() != "some error" {
		t.Error("expected error", err)
	}
}

func TestPublish_WithMultipleSubscribers_CallsDoForEachInSequence(t *testing.T) {
	ctx := context.Background()
	bus := eventbus.New()
	var called []string
	bus.On(testEvent).Do(func(_ context.Context, _ eventbus.Stringer, _ interface{}) error {
		called = append(called, "first")
		return nil
	})
	bus.On(testEvent).Do(func(_ context.Context, _ eventbus.Stringer, _ interface{}) error {
		called = append(called, "second")
		return nil
	})
	bus.On(testEvent).Do(func(_ context.Context, _ eventbus.Stringer, _ interface{}) error {
		called = append(called, "third")
		return nil
	})

	if err := bus.Publish(ctx, testEvent, nil); err != nil {
		t.Error("expected no error", err)
	}

	bus.Flush(ctx)
	if len(called) != 3 {
		t.Error("expected Do to be called 3 times")
	}
}

func TestPublish_WithOneSubscriberThatReturnsError_DoesNotCallOtherSubscribers(t *testing.T) {
	ctx := context.Background()
	bus := eventbus.New()
	var called []string
	bus.On(testEvent).Do(func(_ context.Context, _ eventbus.Stringer, _ interface{}) error {
		called = append(called, "first")
		return errors.New("some error")
	})
	bus.On(testEvent).Do(func(_ context.Context, _ eventbus.Stringer, _ interface{}) error {
		called = append(called, "second")
		return errors.New("some other error")
	})
	bus.On(testEvent).Do(func(_ context.Context, _ eventbus.Stringer, _ interface{}) error {
		called = append(called, "third")
		return nil
	})

	if err := bus.Publish(ctx, testEvent, nil); err == nil || err.Error() != "some error" {
		t.Error("expected error", err)
	}

	bus.Flush(ctx)
	if len(called) != 1 {
		t.Error("expected Do to be called 1 time")
	}
}

func TestPublish_WithClosedBus_ReturnsError(t *testing.T) {
	ctx := context.Background()
	bus := eventbus.New()
	bus.Close()

	if err := bus.Publish(ctx, testEvent, nil); err != eventbus.ErrBusClosed {
		t.Error("expected ErrBusClosed error", err)
	}
}

func TestPublish_WithClosedBus_DoesNotCallDo(t *testing.T) {
	ctx := context.Background()
	bus := eventbus.New()
	bus.Close()
	called := false
	bus.On(testEvent).Do(func(_ context.Context, _ eventbus.Stringer, _ interface{}) error {
		called = true
		return nil
	})

	if err := bus.Publish(ctx, testEvent, nil); err != eventbus.ErrBusClosed {
		t.Error("expected ErrBusClosed error", err)
	}

	bus.Flush(ctx)
	if called {
		t.Error("expected Do to not be called")
	}
}

func TestPublish_WithDoneContext_ReturnsError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	bus := eventbus.New()
	cancel()

	if err := bus.Publish(ctx, testEvent, nil); err != context.Canceled {
		t.Error("expected context.Canceled error", err)
	}
}

func TestPublish_WithHandlerTimeoutOption_SucceedsWithinTimeout(t *testing.T) {
	ctx := context.Background()
	bus := eventbus.New()
	called := false
	bus.On(testEvent).Do(func(_ context.Context, _ eventbus.Stringer, _ interface{}) error {
		time.Sleep(10 * time.Millisecond)
		called = true
		return nil
	})

	if err := bus.Publish(ctx, testEvent, nil, eventbus.WithHandlerTimeoutEventOpt(20*time.Millisecond)); err != nil {
		t.Error("expected no error", err)
	}

	bus.Flush(ctx)
	if !called {
		t.Error("expected Do to be called")
	}
}

func TestPublish_WithHandlerTimeoutOption_FailsAfterTimeout(t *testing.T) {
	ctx := context.Background()
	bus := eventbus.New()
	bus.On(testEvent).Do(func(_ context.Context, _ eventbus.Stringer, _ interface{}) error {
		time.Sleep(20 * time.Millisecond)
		return nil
	})

	if err := bus.Publish(ctx, testEvent, nil, eventbus.WithHandlerTimeoutEventOpt(10*time.Millisecond)); err == nil {
		t.Error("expected ErrHandlerTimeout error", err)
	}
}

func TestPublish_WithHandlerTimoutOption_SucceedsForEachHandlerEvenIfOverallTimeExceedsTimeout(t *testing.T) {
	ctx := context.Background()
	bus := eventbus.New()
	var called []string
	bus.On(testEvent).Do(func(_ context.Context, _ eventbus.Stringer, _ interface{}) error {
		time.Sleep(10 * time.Millisecond)
		called = append(called, "first")
		return nil
	})
	bus.On(testEvent).Do(func(_ context.Context, _ eventbus.Stringer, _ interface{}) error {
		time.Sleep(10 * time.Millisecond)
		called = append(called, "second")
		return nil
	})

	if err := bus.Publish(ctx, testEvent, nil, eventbus.WithHandlerTimeoutEventOpt(15*time.Millisecond)); err != nil {
		t.Error("expected no error", err)
	}

	bus.Flush(ctx)
	if len(called) != 2 {
		t.Error("expected Do to be called 2 times")
	}
}

func TestPublish_WithPublishTimeoutOption_FailsIfTimeExceedsTimeout(t *testing.T) {
	ctx := context.Background()
	bus := eventbus.New()
	bus.On(testEvent).Do(func(_ context.Context, _ eventbus.Stringer, _ interface{}) error {
		time.Sleep(20 * time.Millisecond)
		return nil
	})

	if err := bus.Publish(ctx, testEvent, nil, eventbus.WithPublishTimeoutEventOpt(10*time.Millisecond)); err == nil {
		t.Error("expected ErrPublishTimeout error", err)
	}
}

func TestPublish_WithPublishTimeoutOption_SucceedsIfTimeDoesNotExceedTimeout(t *testing.T) {
	ctx := context.Background()
	bus := eventbus.New()
	bus.On(testEvent).Do(func(_ context.Context, _ eventbus.Stringer, _ interface{}) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	if err := bus.Publish(ctx, testEvent, nil, eventbus.WithPublishTimeoutEventOpt(20*time.Millisecond)); err != nil {
		t.Error("expected no error", err)
	}
}

func TestPublish_WithPublishTimeoutOption_FailsIfOverallTimeExceedsTimeout(t *testing.T) {
	ctx := context.Background()
	bus := eventbus.New()
	bus.On(testEvent).Do(func(_ context.Context, _ eventbus.Stringer, _ interface{}) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})
	bus.On(testEvent).Do(func(_ context.Context, _ eventbus.Stringer, _ interface{}) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	if err := bus.Publish(ctx, testEvent, nil, eventbus.WithPublishTimeoutEventOpt(15*time.Millisecond)); err == nil {
		t.Error("expected ErrPublishTimeout error", err)
	}
}
