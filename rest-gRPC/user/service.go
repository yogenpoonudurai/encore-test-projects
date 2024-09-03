package user

import (
	"encore.app/gen/user/v1/userv1connect"
	"net/http"
)

//encore:service
type Service struct {
	routes http.Handler
}

//encore:api public raw path=/user.v1.UserService/*endpoint
func (s *Service) UserService(w http.ResponseWriter, r *http.Request) {
	s.routes.ServeHTTP(w, r)
}

func initService() (*Service, error) {
	userServ := &UserServer{}
	mux := http.NewServeMux()
	path, handler := userv1connect.NewUserServiceHandler(userServ)
	mux.Handle(path, handler)
	return &Service{routes: mux}, nil

}
