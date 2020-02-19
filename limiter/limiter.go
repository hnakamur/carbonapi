package limiter

import (
	"context"
	"errors"

	"github.com/hnakamur/errstack"
	"github.com/hnakamur/ltsvlog/v3"
)

// ServerLimiter provides interface to limit amount of requests
type RealLimiter struct {
	m   map[string]chan struct{}
	cap int
}

// NewServerLimiter creates a limiter for specific servers list.
func NewServerLimiter(servers []string, l int) ServerLimiter {
	if l <= 0 {
		return &NoopLimiter{}
	}

	sl := make(map[string]chan struct{})

	for _, s := range servers {
		sl[s] = make(chan struct{}, l)
	}

	limiter := &RealLimiter{
		m:   sl,
		cap: l,
	}
	return limiter
}

func (sl RealLimiter) Capacity() int {
	return sl.cap
}

// Enter claims one of free slots or blocks until there is one.
func (sl RealLimiter) Enter(ctx context.Context, s string) error {
	if sl.m == nil {
		return nil
	}

	select {
	case sl.m[s] <- struct{}{}:
		return nil
	case <-ctx.Done():
		ltsvlog.Logger.Err(errstack.New("timeout exceeded"))
		return errors.New("timeout exceeded")
	}
}

// Frees a slot in limiter
func (sl RealLimiter) Leave(ctx context.Context, s string) {
	if sl.m == nil {
		return
	}

	<-sl.m[s]
}
