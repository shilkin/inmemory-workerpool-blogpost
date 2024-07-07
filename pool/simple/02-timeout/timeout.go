package _2_timeout

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
}

func (s *UserService) Create(ctx context.Context, name, email string) error {
	// create user in the database
	userID, _ := s.repo.Create(ctx, name, email)

	nonCancelableCtx := context.WithoutCancel(ctx)

	go func() {
		// to avoid goroutine leak we may set guard timeout,
		// this mitigates goroutine leak, but we still spawn
		// extra goroutine for each service request so under
		// some workload we may still have problems
		//
		// in other words we don't control the number of running
		// workers, clients do
		timeoutCtx, cancel := context.WithTimeout(nonCancelableCtx, 500*time.Millisecond)
		defer cancel()

		s.analytics.Send(timeoutCtx, "user created", userID)
	}()

	return nil
}
