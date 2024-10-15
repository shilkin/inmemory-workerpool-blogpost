package service

import (
	"context"
	"testing"
)

type bar struct {
}

func (bar) GetBarID(_ context.Context, _ int) (int, error) {
	return 1, nil
}

type baz struct {
}

func (baz) GetBazID(_ context.Context, _ int) (int, error) {
	return 2, nil
}

func TestFoo(t *testing.T) {
	pool := NewPool(100)
	foo := NewFooService(pool, bar{}, baz{})

	_, _ = foo.Foo(context.Background(), 3)
}
