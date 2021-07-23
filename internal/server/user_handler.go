package server

import (
	"net/http"

	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
	svc "github.com/w-k-s/simple-budget-tracker/pkg/services"
)

func (a *App) RegisterUser(w http.ResponseWriter, req *http.Request) {

	var (
		createUserRequest svc.CreateUserRequest
		err               error
	)

	if err = a.MustDecodeJson(req.Body, &createUserRequest); err != nil {
		a.MustEncodeProblem(w, req, ledger.NewError(ledger.ErrRequestUnmarshallingFailed, "Failed to parse request", err), http.StatusBadRequest)
		return
	}

	var resp svc.CreateUserResponse
	if resp, err = a.UserService.CreateUser(createUserRequest); err != nil {
		a.MustEncodeProblem(w, req, err, http.StatusBadRequest)
		return
	}

	a.MustEncodeJson(w, resp, http.StatusCreated)
}
