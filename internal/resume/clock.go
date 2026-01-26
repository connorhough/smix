package resume

import (
	"context"
	"time"
)

// Clock abstracts time operations to allow testing.
type Clock interface {
	Now() time.Time
	Sleep(ctx context.Context, d time.Duration) error
}

type realClock struct{}

func (c *realClock) Now() time.Time {
	return time.Now()
}

func (c *realClock) Sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// SystemClock is the default clock implementation.
var SystemClock Clock = &realClock{}
