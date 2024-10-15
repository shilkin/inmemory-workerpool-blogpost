package service

import (
	"context"
	"errors"
	"sync"
)

type FooResponse struct {
	BarID int
	BazID int
}

type barClient interface {
	GetBarID(ctx context.Context, id int) (barID int, err error)
}

type bazClient interface {
	GetBazID(ctx context.Context, id int) (bazID int, err error)
}

type pool interface {
	Enqueue(context.Context, func(poolCtx, taskCtx context.Context)) error
}

type Foo struct {
	pool pool
	bar  barClient
	baz  bazClient
	// ...
}

func NewFooService(pool pool, bar barClient, baz bazClient) *Foo {
	return &Foo{
		pool: pool,
		bar:  bar,
		baz:  baz,
	}
}

// H/W:
// (1) implement Foo with channels <-
// (2) stop all tasks when at least one task fails

// GET /api/v1/foo -> json FooResponse
func (f *Foo) Foo(ctx context.Context, id int) (*FooResponse, error) {
	wg := sync.WaitGroup{}

	var barID int
	var bazID int

	var errorBar error
	var errorBaz error

	wg.Add(2)

	// go func() {}
	_ = f.pool.Enqueue(ctx, func(_, _ context.Context) {
		defer wg.Done()

		barID, errorBar = f.bar.GetBarID(ctx, id)
		if errorBar != nil {
			return
		}
	})

	// go func() {}
	_ = f.pool.Enqueue(ctx, func(_, _ context.Context) {
		defer wg.Done()

		bazID, errorBaz = f.baz.GetBazID(ctx, id)
		if errorBaz != nil {
			return
		}
	})

	wg.Wait() // block

	if errorBar != nil || errorBaz != nil {
		return nil, errors.Join(errorBar, errorBaz)
	}

	return &FooResponse{
		BarID: barID,
		BazID: bazID,
	}, nil
}
