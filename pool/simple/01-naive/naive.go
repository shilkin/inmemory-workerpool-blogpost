package _1_naive

import "context"

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

	// send analytics event asynchronously
	//
	// IMPORTANT (1)
	// note that we disable context cancellation
	// for this goroutine to avoid cascading context
	// cancelling; prior to go1.21 you have to
	// write your own un-canceling function
	//
	// IMPORTANT (2)
	// this doesn't affect business critical flow
	// but this creates potential goroutine leak if analytics
	// is works slowly or even hangs
	go s.analytics.Send(context.WithoutCancel(ctx), "user created", userID)

	return nil
}
