package eventbus

import (
	"context"
	"time"
)

type (
	observer interface {
		Observe(ctx context.Context, name Stringer, data interface{})
	}
	observerWithOptions struct {
		observer
		opts observerOptions
	}
	observerOptions struct {
		timeout time.Duration
	}
)
