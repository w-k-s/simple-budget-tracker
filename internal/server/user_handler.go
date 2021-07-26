package server

import (
	"net/http"

	svc "github.com/w-k-s/simple-budget-tracker/pkg/services"
)

func (a *App) RegisterUser(w http.ResponseWriter, req *http.Request) {

	var (
		createUserRequest svc.CreateUserRequest
		resp              svc.CreateUserResponse
		err               error
	)

	if ok := a.DecodeJsonOrSendBadRequest(w, req, &createUserRequest); !ok {
		return
	}

	if resp, err = a.UserService.CreateUser(createUserRequest); err != nil {
		a.MustEncodeProblem(w, req, err)
		return
	}

	a.MustEncodeJson(w, resp, http.StatusCreated)
}
