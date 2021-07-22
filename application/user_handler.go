package application

import (
	"net/http"

	"github.com/w-k-s/simple-budget-tracker/core"
)

func (a *App) RegisterUser(w http.ResponseWriter, req *http.Request) {

	var (
		createUserRequest core.CreateUserRequest
		err               error
	)

	if err = a.MustDecodeJson(req.Body, &createUserRequest); err != nil {
		a.MustEncodeProblem(w, req, core.NewError(core.ErrRequestUnmarshallingFailed, "Failed to parse request", err), http.StatusBadRequest)
		return
	}

	var resp core.CreateUserResponse
	if resp, err = a.UserService.CreateUser(createUserRequest); err != nil {
		a.MustEncodeProblem(w, req, err, http.StatusBadRequest)
		return
	}

	a.MustEncodeJson(w, resp, http.StatusCreated)
}
