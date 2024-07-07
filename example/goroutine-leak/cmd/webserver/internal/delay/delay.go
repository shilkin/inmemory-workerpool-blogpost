package delay

import (
	"context"
	"time"
)

const defaultDelay = time.Millisecond

type delayCtxKey struct{}

func ToContext(ctx context.Context, delay time.Duration) context.Context {
	return context.WithValue(ctx, delayCtxKey{}, delay)
}

func FromContext(ctx context.Context) time.Duration {
	if delay, ok := ctx.Value(delayCtxKey{}).(time.Duration); ok {
		return delay
	}

	return defaultDelay
}
