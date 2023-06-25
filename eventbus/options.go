package eventbus

import "time"

type (
	eventOpt    func(*Event)
	busOpt      func(*bus)
	observerOpt func(*observerOptions)
)

// Bus options
var (
	WithMaxConcurrencyBusOpt = func(c int64) busOpt {
		return func(b *bus) {
			if c < 1 {
				c = 1
			}
			b.concurrency = c
		}
	}
	WithContinueOnErrorBusOpt = func() busOpt {
		return func(b *bus) {
			b.continueOnError = true
		}
	}
)

// Event options
var (
	WithHandlerTimeoutEventOpt = func(d time.Duration) eventOpt {
		return func(e *Event) {
			e.handlerTimeout = d
		}
	}
	WithPublishTimeoutEventOpt = func(d time.Duration) eventOpt {
		return func(e *Event) {
			e.publishTimeout = d
		}
	}
)

// Observer options
var (
	WithTimeoutObserverOpt = func(d time.Duration) observerOpt {
		return func(o *observerOptions) {
			o.timeout = d
		}
	}
)
