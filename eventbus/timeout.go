package eventbus

import (
	"context"
	"time"
)

// doWithTimeout executes a function with a context and a timeout.
// The provided function is run in a separate goroutine, and doWithTimeout
// waits for either the function to complete or for the context to be canceled
// or the timeout to elapse, whichever happens first.
//
// Parameters:
//   - ctx: The parent context. The function respects the cancellation or
//     deadline of this context. If ctx is already canceled or past its
//     deadline, doWithTimeout returns immediately with an error.
//   - timeout: The maximum duration to wait for the function to complete.
//     If timeout is positive, a new context with this timeout is
//     derived from ctx and passed to the function. If it is zero or
//     negative, the function is executed with the original ctx.
//   - fn: The function to be executed. It should take a context as a parameter.
//     The context passed to this function is canceled if the timeout
//     elapses or the parent context is canceled, so the function can
//     use this context to release resources or abort early.
//
// Returns:
//   - If the function completes before the context is canceled or the timeout
//     elapses, the error returned by the function (or nil) is returned.
//   - If the context is canceled or the timeout elapses before the function
//     completes, an error indicating that the context was canceled is returned.
//
// Note:
//   - If the function takes a long time to execute, this function will block
//     for that duration or until the context is canceled or the timeout
//     is reached.
//   - If the context's timeout elapses before the function has finished executing,
//     the goroutine running the function will keep running until it's done.
func doWithTimeout(ctx context.Context, timeout time.Duration, fn func(context.Context) error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	done := make(chan error)
	go func() {
		done <- fn(ctx)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

// shortestDuration takes a variadic number of time.Duration values and returns the
// shortest duration among them. If no durations are passed, it returns 0.
//
// Parameters:
//   - durations: A variadic number of time.Duration values.
//
// Returns:
//   - The shortest duration among the passed durations, or 0 if no durations
//     were passed.
//
// Example:
//
//	shortest := shortestDuration(time.Second, 2*time.Second, 500*time.Millisecond)
//	fmt.Println(shortest) // Output: 500ms
func shortestDuration(durations ...time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0 // or return a special value to indicate no durations were passed
	}

	shortest := durations[0]
	for _, d := range durations[1:] { // Start iterating from index 1
		if d < shortest {
			shortest = d
		}
	}
	return shortest
}
