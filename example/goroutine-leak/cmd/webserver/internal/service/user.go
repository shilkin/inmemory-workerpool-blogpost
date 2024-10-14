package service

import (
	"context"
	"fmt"
)

// type WorkerPool interface {
// 	// Enqueue enqueues new task
// 	Enqueue(ctx context.Context, task func(ctx context.Context)) error

// 	// Stop gracefully stops the pool
// 	// Stop(/*ctx?*/)
// }

type UserRepository interface {
	Create(ctx context.Context, name, email string) (string, error)
}

type Analytics interface {
	Send(ctx context.Context, message string, args ...string)
}

type UserService struct {
	repo       UserRepository
	analytics  Analytics
	sema       chan struct{}
	pool       WorkerPool
	barService barService
	bazService bazService
}

func NewUserService(repo UserRepository, analytics Analytics) *UserService {
	return &UserService{
		repo:      repo,
		analytics: analytics,
		sema:      make(chan struct{}, 100),
	}
}

// https://pkg.go.dev/golang.org/x/sync/semaphore
func (s *UserService) Create(ctx context.Context, name, email string) error {
	// create user in the database
	userID, _ := s.repo.Create(ctx, name, email)

	err := s.pool.Enqueue(ctx, func(poolCtx, taskCtx context.Context) {
		s.analytics.Send(taskCtx, "user created", userID)
	})
	if err != nil {
		return fmt.Errorf("enqueue task: %w", err)
	}

	return nil
}

type FooResult struct {
	BarID int
	BazID int
}

type serviceResult struct {
	id  int
	err error
}

type barService interface {
	GetBarID(ctx context.Context, id int) (int, error)
}

type bazService interface {
	GetBazID(ctx context.Context, id int) (int, error)
}

// GET /api/v1/foo -> json: FooResult
//
// H/W: (3) think how to utilise mutex to simplify concurrency code: Foo1() {}
func (s *UserService) Foo(ctx context.Context, fooID int) (*FooResult, error) {
	barChan := make(chan serviceResult)

	// H/W: (1) refactor task creation in order to simplify the code
	err := s.pool.Enqueue(ctx, func(_, _ context.Context) {
		id, err := s.barService.GetBarID(ctx, fooID)

		select {
		case barChan <- serviceResult{id: id, err: err}:
		case <-ctx.Done():
			return
		}

	})
	if err != nil {
		return nil, fmt.Errorf("enqueue bar task: %w", err)
	}

	bazChan := make(chan serviceResult)

	err = s.pool.Enqueue(ctx, func(_, _ context.Context) {
		id, err := s.bazService.GetBazID(ctx, fooID)

		select {
		case bazChan <- serviceResult{id: id, err: err}:
		case <-ctx.Done():
			return
		}

	})
	if err != nil {
		return nil, fmt.Errorf("enqueue baz task: %w", err)
	}

	var result FooResult

	// H/W: (2) simplify waiting the results
	select {
	case r := <-barChan:
		if r.err != nil {
			return nil, fmt.Errorf("error bar task: %w", r.err)
		}

		result.BarID = r.id
	case <-ctx.Done():
		return nil, fmt.Errorf("waiting bar result: %w", ctx.Err())
	}

	select {
	case r := <-bazChan:
		if r.err != nil {
			return nil, fmt.Errorf("error baz task: %w", r.err)
		}

		result.BazID = r.id
	case <-ctx.Done():
		return nil, fmt.Errorf("waiting baz result: %w", ctx.Err())
	}

	return &result, nil
}
