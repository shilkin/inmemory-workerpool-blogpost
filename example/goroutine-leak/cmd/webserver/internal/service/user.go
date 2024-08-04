package service

import (
	"context"
	"errors"
	"sync"
)

// 1. Graceful shutdown
// 2. *Pass singal about pool stopping to task handlers
// 3. General purpose pool
// type Task func(ctx context.Context) {

// }

type Task struct {
	f  func(ctx context.Context)
	id int
}

type Pool interface {
	Enqueue(ctx context.Context, task Task) (int, error)
	Stop(taskId int) error
}

type PoolService struct {
	ms         sync.Mutex
	tasksQueue []*Task
	counter    int
	workers    map[int]func()
	maxWorkers int
}

func (ps *PoolService) Enqueue(ctx context.Context, task Task) (int, error) {
	ps.ms.Lock()
	ps.counter = ps.counter + 1
	task.id = ps.counter
	ps.tasksQueue = append(ps.tasksQueue, &task)
	ps.ms.Unlock()
	return task.id, nil
}

func (ps *PoolService) Stop(taskId int) error {
	if ps.workers[taskId] != nil {
		ps.workers[taskId]()
	} else {
		return errors.New("wrong taskId")
	}
	return nil
}

func (ps *PoolService) Run() error {
	for {
		if len(ps.tasksQueue) > 0 && len(ps.workers) < ps.maxWorkers {
			ctx, cancel := context.WithCancel(context.Background())
			ps.ms.Lock()
			ps.workers[ps.tasksQueue[0].id] = cancel
			fn := func() {
				defer delete(ps.workers, ps.tasksQueue[0].id)
				ps.tasksQueue[0].f(ctx)
			}
			go fn()
			ps.tasksQueue = ps.tasksQueue[:1]
			ps.ms.Unlock()
		}
	}
}

type UserRepository interface {
	Create(ctx context.Context, name, email string) (string, error)
}

type Analytics interface {
	Send(ctx context.Context, message string, args ...string)
}

type UserService struct {
	repo       UserRepository
	analytics  Analytics
	poolSevice Pool
	poolC      chan struct{}
}

func NewUserService(repo UserRepository, analytics Analytics) *UserService {
	poolService := PoolService{maxWorkers: 10}
	go poolService.Run()
	return &UserService{
		repo:       repo,
		analytics:  analytics,
		poolSevice: &poolService,
		poolC:      make(chan struct{}, 10),
	}
}

func (s *UserService) Create(ctx context.Context, name, email string) error {
	// create user in the database
	userID, _ := s.repo.Create(ctx, name, email)
	// s.analytics.Send(ctx, "user created", userID)

	s.poolSevice.Enqueue(context.WithoutCancel(ctx), Task{f: func(ctx context.Context) {
		s.analytics.Send(context.WithoutCancel(ctx), "user created", userID)
	}})

	// select {
	// case s.poolC <- struct{}{}: // acquire worker
	// case <-ctx.Done():
	// 	return ctx.Err() // out of resources
	// default:

	// fallback
	// go s.analytics.Send(ctx, "user created", userID)
	// }

	// go func() {
	// 	defer func() { <-s.poolC }() // release worker

	// s.analytics.Send(context.WithoutCancel(ctx), "user created", userID)
	// }()

	return nil
}
