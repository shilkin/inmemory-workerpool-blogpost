package service

import (
	"context"
	"errors"
	"fmt"
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

type MyFuncType func(ctx context.Context, id int) (barID int, err error)

func (f *Foo) someFunc(con context.Context, stopCon context.Context, cancelFunc context.CancelFunc, getFunc MyFuncType, id int, result *int, err *error) chan struct{} {
	finish := make(chan struct{})
	_ = f.pool.Enqueue(con, func(_, _ context.Context) {

		defer func() {
			finish <- struct{}{}
		}()

		res, respErr := getFunc(con, id)
		if respErr != nil {
			err = &respErr
			cancelFunc()
			return
		}

		result = &res
	})

	select {
	case <-finish:
		return finish
	case <-stopCon.Done():
		return finish
	}
}

// H/W:
// (1) implement Foo with channels <-
// (2) stop all tasks when at least one task fails

// GET /api/v1/foo -> json FooResponse
func (f *Foo) Foo(ctx context.Context, id int) (*FooResponse, error) {
	defContext, cancel := context.WithCancel(ctx)

	var barID int
	var bazID int

	var errorBar error
	var errorBaz error

	a := f.someFunc(ctx, defContext, cancel, f.bar.GetBarID, id, &barID, &errorBar)
	b := f.someFunc(ctx, defContext, cancel, f.baz.GetBazID, id, &bazID, &errorBaz)
	d, e := <-a, <-b
	fmt.Print(d, e)

	if errorBar != nil || errorBaz != nil {
		return nil, errors.Join(errorBar, errorBaz)
	}

	return &FooResponse{
		BarID: barID,
		BazID: bazID,
	}, nil
}
