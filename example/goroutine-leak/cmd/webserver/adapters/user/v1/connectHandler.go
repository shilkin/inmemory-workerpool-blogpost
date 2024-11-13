package user_v1

import (
	"connectrpc.com/connect"
	"context"
	connect2 "github.com/shilkin/inmemory-workerpool-blogpost/example/goroutine-leak/cmd/webserver/gen/user/models"
	"github.com/shilkin/inmemory-workerpool-blogpost/example/goroutine-leak/cmd/webserver/internal/service"
)

type userConnectHandler struct {
	userService *service.UserService
}

func NewUserConnectHandler(userService *service.UserService) *userConnectHandler {
	return &userConnectHandler{userService: userService}
}

func (s *userConnectHandler) CreateUser(
	ctx context.Context,
	req *connect.Request[connect2.CreateReq],
) (*connect.Response[connect2.CreateResp], error) {
	name := req.Msg.Name
	email := req.Msg.Email

	err := s.userService.Create(ctx, name, email)
	if err != nil {
		resp := &connect2.CreateResp{
			Error: true,
			Msg:   err.Error(),
		}
		return connect.NewResponse(resp), err
	}

	resp := &connect2.CreateResp{
		Error: false,
		Msg:   "User created successfully",
	}
	return connect.NewResponse(resp), nil
}
