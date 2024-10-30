package service_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/shilkin/inmemory-workerpool-blogpost/example/goroutine-leak/cmd/webserver/internal/service"
	mock_service "github.com/shilkin/inmemory-workerpool-blogpost/example/goroutine-leak/cmd/webserver/internal/service/internal/mock"
	"github.com/stretchr/testify/require"
)

func TestUserServiceCreate(t *testing.T) {
	// happy path
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish) // defer ctrl.Finish() <- linter warning

	repoMock := mock_service.NewMockUserRepository(ctrl)
	analyticsMock := mock_service.NewMockAnalytics(ctrl)
	userService := service.NewUserService(repoMock, analyticsMock)

	var isAnalyticsCalled atomic.Bool // false

	repoMock.EXPECT().Create(context.Background(), "Jon Doe", "jon.doe@example.com").Return("userID", nil)
	analyticsMock.EXPECT().Send(gomock.Any(), "user created", "userID").Do(func(_ context.Context, _, _ string) {
		isAnalyticsCalled.Store(true)
	})

	err := userService.Create(context.Background(), "Jon Doe", "jon.doe@example.com")

	// wait for analytics being called eventually in 1s or fail
	require.Eventually(t, isAnalyticsCalled.Load, time.Second, time.Millisecond) // sleep

	require.NoError(t, err)
}
