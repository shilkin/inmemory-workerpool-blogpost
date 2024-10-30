package service_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/shilkin/inmemory-workerpool-blogpost/example/goroutine-leak/cmd/webserver/internal/service"
	mock "github.com/shilkin/inmemory-workerpool-blogpost/example/goroutine-leak/cmd/webserver/internal/service/internal/mock"
	"github.com/stretchr/testify/require"
)

func TestUserCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish) // defer ctrl.Finish() <- linters warning

	repoMock := mock.NewMockUserRepository(ctrl)
	analyticsMock := mock.NewMockAnalytics(ctrl)

	var isAnalyticsCalled atomic.Bool

	repoMock.EXPECT().Create(context.Background(), "Jon Doe", "jon.doe@example.com").Return("id", nil)
	analyticsMock.EXPECT().Send(gomock.Any(), "user created", "id").
		Do(func(_ context.Context, _, _ string) {
			isAnalyticsCalled.Store(true)
		})

	userService := service.NewUserService(repoMock, analyticsMock)

	err := userService.Create(context.Background(), "Jon Doe", "jon.doe@example.com")
	require.NoError(t, err) // fail now

	// wait for analytics being called or fail the test
	require.Eventually(t, isAnalyticsCalled.Load, time.Second, time.Millisecond)
}
