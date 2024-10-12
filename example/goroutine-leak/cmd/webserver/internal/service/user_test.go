package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/shilkin/inmemory-workerpool-blogpost/example/goroutine-leak/cmd/webserver/internal/service"
	mock "github.com/shilkin/inmemory-workerpool-blogpost/example/goroutine-leak/cmd/webserver/internal/service/internal/mock"
	"github.com/stretchr/testify/require"
)

/*
ok  	github.com/shilkin/inmemory-workerpool-blogpost/example/goroutine-leak/cmd/webserver/internal/service	0.836s
*/

func TestUserCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish() // t.Cleanup(ctrl.Finish)

	repoMock := mock.NewMockUserRepository(ctrl)
	analyticsMock := mock.NewMockAnalytics(ctrl)
	poolMock := mock.NewMockPool(ctrl)

	userService := service.NewUserService(repoMock, analyticsMock, poolMock)

	ctx := context.Background()

	t.Run("user created succesfully", func(t *testing.T) {
		gomock.InOrder(
			poolMock.EXPECT().Enqueue(ctx, gomock.Any()).Return(nil),
			repoMock.EXPECT().Create(ctx, "test", "test@test.test").Return("userID", nil),
		)

		err := userService.Create(ctx, "test", "test@test.test")
		require.NoError(t, err)
	})

	t.Run("task enqueueing failed", func(t *testing.T) {
		repoMock.EXPECT().Create(ctx, "test", "test@test.test").Return("userID", nil)
		poolMock.EXPECT().Enqueue(ctx, gomock.Any()).Return(errors.New("enqueue error"))
		err := userService.Create(ctx, "test", "test@test.test")
		require.Error(t, err)
	})

	t.Run("repository saving failed", func(t *testing.T) {
		repoMock.EXPECT().Create(ctx, "test", "test@test.test").Return("userID", errors.New("create error"))
		err := userService.Create(ctx, "test", "test@test.test")
		require.Error(t, err)
	})
}
