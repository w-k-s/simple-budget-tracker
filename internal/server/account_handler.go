package server

import (
	"net/http"

	svc "github.com/w-k-s/simple-budget-tracker/pkg/services"
)

func (a *App) RegisterAccounts(w http.ResponseWriter, req *http.Request) {

	var (
		createAccountsRequest svc.CreateAccountsRequest
		resp                  svc.AccountsResponse
		err                   error
	)

	if ok := a.DecodeJsonOrSendBadRequest(w, req, &createAccountsRequest); !ok {
		return
	}

	if resp, err = a.AccountService.CreateAccounts(req.Context(), createAccountsRequest); err != nil {
		a.MustEncodeProblem(w, req, err)
		return
	}

	a.MustEncodeJson(w, resp, http.StatusCreated)
}

func (a *App) GetAccounts(w http.ResponseWriter, req *http.Request) {
	var (
		resp svc.AccountsResponse
		err  error
	)

	if resp, err = a.AccountService.GetAccounts(req.Context()); err != nil {
		a.MustEncodeProblem(w, req, err)
		return
	}

	a.MustEncodeJson(w, resp, http.StatusOK)
}
