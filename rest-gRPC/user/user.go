package user

import (
	"connectrpc.com/connect"
	userv1 "encore.app/gen/user/v1"
	"golang.org/x/net/context"
	"log"
)

type UserServer struct{}

func (s *UserServer) Get(
	ctx context.Context,
	req *connect.Request[userv1.UserRequest],
) (*connect.Response[userv1.UserResponse], error) {
	log.Println("Request headers: ", req.Header())
	res := connect.NewResponse(&userv1.UserResponse{
		Id: "1",
	})
	res.Header().Set("Greet-Version", "v1")
	return res, nil
}
