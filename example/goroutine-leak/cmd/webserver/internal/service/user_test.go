package service_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/shilkin/inmemory-workerpool-blogpost/example/goroutine-leak/cmd/webserver/internal/service"
	mock "github.com/shilkin/inmemory-workerpool-blogpost/example/goroutine-leak/cmd/webserver/internal/service/internal/mock"
	"github.com/stretchr/testify/require"
)

func TestUserCreate(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish) // defer ctrl.Finish() <- linters warning

	t.Run("user created sucessfully with real pool", func(t *testing.T) {
		t.Parallel()

		repoMock := mock.NewMockUserRepository(ctrl)
		analyticsMock := mock.NewMockAnalytics(ctrl)
		pool := service.NewPool(1) // real pool

		var isAnalyticsCalled atomic.Bool

		repoMock.EXPECT().Create(context.Background(), "Jon Doe", "jon.doe@example.com").Return("id", nil)
		analyticsMock.EXPECT().Send(gomock.Any(), "user created", "id").
			Do(func(_ context.Context, _, _ string) {
				isAnalyticsCalled.Store(true)
			})

		userService := service.NewUserService(repoMock, analyticsMock, pool)

		err := userService.Create(context.Background(), "Jon Doe", "jon.doe@example.com")
		require.NoError(t, err) // fail now

		// wait for analytics being called or fail the test
		require.Eventually(t, isAnalyticsCalled.Load, time.Second, time.Millisecond)
	})
	t.Run("user created sucessfully", func(t *testing.T) {
		t.Parallel()

		repoMock := mock.NewMockUserRepository(ctrl)
		analyticsMock := mock.NewMockAnalytics(ctrl)
		pooMock := mock.NewMockPool(ctrl)

		repoMock.EXPECT().Create(context.Background(), "Jon Doe", "jon.doe@example.com").Return("id", nil)
		pooMock.EXPECT().Enqueue(context.Background(), gomock.Any()).Return(nil)

		userService := service.NewUserService(repoMock, analyticsMock, pooMock)

		err := userService.Create(context.Background(), "Jon Doe", "jon.doe@example.com")
		require.NoError(t, err) // fail now
	})

	t.Run("unable to save user in the repo", func(t *testing.T) {
		t.Parallel()

		repoMock := mock.NewMockUserRepository(ctrl)
		analyticsMock := mock.NewMockAnalytics(ctrl)
		pooMock := mock.NewMockPool(ctrl)

		repoMock.EXPECT().Create(context.Background(), "Jon Doe", "jon.doe@example.com").Return("1", errors.New("Except"))

		userService := service.NewUserService(repoMock, analyticsMock, pooMock)

		err := userService.Create(context.Background(), "Jon Doe", "jon.doe@example.com")

		require.Error(t, err) // fail now
		require.Contains(t, err.Error(), "Except")
	})

	t.Run("unable to enqueue job", func(t *testing.T) {
		t.Parallel()

		repoMock := mock.NewMockUserRepository(ctrl)
		analyticsMock := mock.NewMockAnalytics(ctrl)
		pooMock := mock.NewMockPool(ctrl)

		gomock.InOrder(
			repoMock.EXPECT().Create(context.Background(), "Jon Doe", "jon.doe@example.com").Return("1", nil),
			pooMock.EXPECT().Enqueue(context.Background(), gomock.Any()).Return(errors.New("Enqueue Except")),
		)

		userService := service.NewUserService(repoMock, analyticsMock, pooMock)

		err := userService.Create(context.Background(), "Jon Doe", "jon.doe@example.com")

		require.Error(t, err) // fail now
		require.Contains(t, err.Error(), "Enqueue Except")
	})
}
