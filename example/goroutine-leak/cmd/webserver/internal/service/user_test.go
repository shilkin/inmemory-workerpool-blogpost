package service_test

import (
	"context"
	"errors"
	"github.com/stretchr/testify/suite"
	"testing"

	"github.com/golang/mock/gomock"
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
