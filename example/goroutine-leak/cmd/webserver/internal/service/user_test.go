package service_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/golang/mock/gomock"
	pool "github.com/shilkin/inmemory-workerpool-blogpost/example/goroutine-leak/cmd/webserver/internal/pool"
	"github.com/shilkin/inmemory-workerpool-blogpost/example/goroutine-leak/cmd/webserver/internal/service"
	mock "github.com/shilkin/inmemory-workerpool-blogpost/example/goroutine-leak/cmd/webserver/internal/service/internal/mock"
	"github.com/stretchr/testify/require"
)

type UserSuite struct {
	suite.Suite
	ctrl          *gomock.Controller
	repoMock      *mock.MockUserRepository
	analyticsMock *mock.MockAnalytics
	poolMock      *mock.MockPool
	userService   *service.UserService
	ctx           context.Context
}

func (s *UserSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.repoMock = mock.NewMockUserRepository(s.ctrl)
	s.analyticsMock = mock.NewMockAnalytics(s.ctrl)
	s.poolMock = mock.NewMockPool(s.ctrl)
	s.userService = service.NewUserService(s.repoMock, s.analyticsMock, s.poolMock)
	s.ctx = context.Background()
}

func (s *UserSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *UserSuite) TestUserCreate_Success() {
	gomock.InOrder(
		s.repoMock.EXPECT().Create(s.ctx, "test", "test@test.test").Return("userID", nil),
		s.poolMock.EXPECT().Enqueue(s.ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, task func(ctx context.Context)) error {
				task(ctx)
				return nil
			},
		),
		s.analyticsMock.EXPECT().Send(gomock.Any(), "user created", "userID"),
	)

	err := s.userService.Create(s.ctx, "test", "test@test.test")
	require.NoError(s.T(), err)
}

func (s *UserSuite) TestUserCreate_TaskEnqueueFailed() {
	gomock.InOrder(
		s.repoMock.EXPECT().Create(s.ctx, "test", "test@test.test").Return("userID", nil),
		s.poolMock.EXPECT().Enqueue(s.ctx, gomock.Any()).Return(errors.New("enqueue error")),
	)

	err := s.userService.Create(s.ctx, "test", "test@test.test")
	require.Error(s.T(), err)
}

func (s *UserSuite) TestUserCreate_RepositorySavingFailed() {
	gomock.InOrder(
		s.repoMock.EXPECT().Create(s.ctx, "test", "test@test.test").Return("userID", errors.New("create error")),
	)

	err := s.userService.Create(s.ctx, "test", "test@test.test")
	require.Error(s.T(), err)
}

func TestUserSuite(t *testing.T) {
	suite.Run(t, new(UserSuite))
}

func TestWithRealPool(t *testing.T) {
	ctrl := gomock.NewController(t)
	repoMock := mock.NewMockUserRepository(ctrl)
	analyticsMock := mock.NewMockAnalytics(ctrl)
	pool := pool.NewPoolService(10)
	userService := service.NewUserService(repoMock, analyticsMock, pool)
	ctx := context.Background()

	var isAnalyticsCalled atomic.Bool

	gomock.InOrder(
		repoMock.EXPECT().Create(ctx, "test", "test@test.test").Return("userID", nil),
		analyticsMock.EXPECT().Send(gomock.Any(), "user created", "userID").
			Do(func(_ context.Context, _, _ string) {
				isAnalyticsCalled.Store(true)
			}),
	)

	err := userService.Create(ctx, "test", "test@test.test")

	// time.Sleep(1 * time.Second)
	// runtime.Gosched()

	// calls condition every 1ms not longer than 1s then fails if condition is false
	// success otherwise
	assert.Eventually(t, isAnalyticsCalled.Load, time.Second, time.Millisecond)

	require.NoError(t, err)

}
