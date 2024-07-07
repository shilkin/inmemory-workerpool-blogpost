package _0_nopool

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

	// send analytics event synchronously
	//
	// if analytics works slowly then this auxiliary
	// logic affects user creation business critical flow
	s.analytics.Send(ctx, "user created", userID)

	return nil
}
