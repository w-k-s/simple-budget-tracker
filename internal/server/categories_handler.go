package server

import (
	"net/http"

	svc "github.com/w-k-s/simple-budget-tracker/pkg/services"
)

func (a *App) CreateCategories(w http.ResponseWriter, req *http.Request) {

	var (
		createCategoriesRequest svc.CreateCategoriesRequest
		resp                    svc.CategoriesResponse
		err                     error
	)

	if ok := a.DecodeJsonOrSendBadRequest(w, req, &createCategoriesRequest); !ok {
		return
	}

	if resp, err = a.CategoriesService.CreateCategories(req.Context(), createCategoriesRequest); err != nil {
		a.MustEncodeProblem(w, req, err)
		return
	}

	a.MustEncodeJson(w, resp, http.StatusCreated)
}

func (a *App) GetCategories(w http.ResponseWriter, req *http.Request) {
	var (
		resp svc.CategoriesResponse
		err  error
	)

	if resp, err = a.CategoriesService.GetCategories(req.Context()); err != nil {
		a.MustEncodeProblem(w, req, err)
		return
	}

	a.MustEncodeJson(w, resp, http.StatusOK)
}
