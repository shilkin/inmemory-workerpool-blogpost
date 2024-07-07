package _3_pool

import (
	"context"
	"time"
)

type UserRepository interface {
	Create(ctx context.Context, name, email string) (string, error)
}

type Analytics interface {
	Send(ctx context.Context, message string, args ...string)
}

type UserService struct {
	repo      UserRepository
	analytics Analytics

	// poolC is a buffered channel for using as semaphore;
	//
	// buffer size is a semaphore max counter size.
	poolC chan struct{}
}

func (s *UserService) Create(ctx context.Context, name, email string) error {
	// create user in the database
	userID, _ := s.repo.Create(ctx, name, email)

	select {
	// request timeout
	case <-ctx.Done():
		return ctx.Err()
	// acquire worker
	case s.poolC <- struct{}{}:
		// release worker
		defer func() { <-s.poolC }()
	// if all workers are busy
	// fallback to synchronous mode
	default:
		// optionally we may fail here
		// return errors.New("no workers available")

		// no matter how the parent request timeout is
		// we also set guard timeout here
		timeoutCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
		defer cancel()

		s.analytics.Send(timeoutCtx, "user created", userID)

		return nil
	}

	nonCancelableCtx := context.WithoutCancel(ctx)

	// send analytics event asynchronously
	go func() {
		// don't forget about guard timeout
		timeoutCtx, cancel := context.WithTimeout(nonCancelableCtx, 500*time.Millisecond)
		defer cancel()

		s.analytics.Send(timeoutCtx, "user created", userID)
	}()

	return nil
}
