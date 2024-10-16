package service

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

func NewUserService(repo UserRepository, analytics Analytics) *UserService {
	return &UserService{repo: repo, analytics: analytics}
}

func (s *UserService) Create(ctx context.Context, name, email string) error {
	// create user in the database
	userID, _ := s.repo.Create(ctx, name, email)

	ctx = context.WithoutCancel(ctx)

	// send analytics event synchronously
	// which may cause goroutine and memory leak
	go s.analytics.Send(ctx, "user created", userID)

	return nil
}
