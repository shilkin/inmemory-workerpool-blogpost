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
	"github.com/stretchr/testify/suite"
)

type TestUserCreateSuite struct {
	suite.Suite
	repoMock      *mock.MockUserRepository
	analyticsMock *mock.MockAnalytics
	pooMock       *mock.MockPool
}

func (s *TestUserCreateSuite) SetupTest() {
	s.T().Parallel()
	ctrl := gomock.NewController(s.T())
	s.T().Cleanup(ctrl.Finish)

	s.repoMock = mock.NewMockUserRepository(ctrl)
	s.analyticsMock = mock.NewMockAnalytics(ctrl)
	s.pooMock = mock.NewMockPool(ctrl)
}

func (s *TestUserCreateSuite) TestUserCreateSuccessWithPool() {
	var isAnalyticsCalled atomic.Bool

	s.repoMock.EXPECT().Create(context.Background(), "Jon Doe", "jon.doe@example.com").Return("id", nil)
	s.analyticsMock.EXPECT().Send(gomock.Any(), "user created", "id").
		Do(func(_ context.Context, _, _ string) {
			isAnalyticsCalled.Store(true)
		})

	userService := service.NewUserService(s.repoMock, s.analyticsMock, s.pooMock)

	err := userService.Create(context.Background(), "Jon Doe", "jon.doe@example.com")
	s.Require().NoError(err)
	s.Require().Eventually(func() bool {
		return isAnalyticsCalled.Load()
	}, time.Second, time.Millisecond)
}

func (s *TestUserCreateSuite) TestUserCreateSuccessfully() {
	s.repoMock.EXPECT().Create(context.Background(), "Jon Doe", "jon.doe@example.com").Return("id", nil)
	s.pooMock.EXPECT().Enqueue(context.Background(), gomock.Any()).Return(nil)

	userService := service.NewUserService(s.repoMock, s.analyticsMock, s.pooMock)

	err := userService.Create(context.Background(), "Jon Doe", "jon.doe@example.com")
	s.Require().NoError(err)
}

func (s *TestUserCreateSuite) TestUnableToSaveUserInTheRepo() {
	s.repoMock.EXPECT().Create(context.Background(), "Jon Doe", "jon.doe@example.com").Return("1", errors.New("Except"))

	userService := service.NewUserService(s.repoMock, s.analyticsMock, s.pooMock)

	err := userService.Create(context.Background(), "Jon Doe", "jon.doe@example.com")

	s.Require().Error(err)
	s.Require().Contains(s.T(), err, "Except")
}

func (s *TestUserCreateSuite) TestUnableToSendAnalytics() {
	gomock.InOrder(
		s.repoMock.EXPECT().Create(context.Background(), "Jon Doe", "jon.doe@example.com").Return("1", nil),
		s.pooMock.EXPECT().Enqueue(context.Background(), gomock.Any()).Return(errors.New("Enqueue Except")),
	)

	userService := service.NewUserService(s.repoMock, s.analyticsMock, s.pooMock)

	err := userService.Create(context.Background(), "Jon Doe", "jon.doe@example.com")

	s.Require().Error(err)
	s.Require().Contains(s.T(), err, "Enqueue Except")

}

func TestUserCreate(t *testing.T) {
	suite.Run(t, new(TestUserCreateSuite))
}
