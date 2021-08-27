package server

import (
	"log"
	"net/http"
	"strconv"

	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
	svc "github.com/w-k-s/simple-budget-tracker/pkg/services"
)

func (a *App) CreateRecord(w http.ResponseWriter, req *http.Request) {

	var (
		createRecordRequest svc.CreateRecordRequest
		resp                svc.RecordResponse
		err                 error
	)

	if ok := a.DecodeJsonOrSendBadRequest(w, req, &createRecordRequest); !ok {
		return
	}

	if resp, err = a.RecordService.CreateRecord(req.Context(), createRecordRequest); err != nil {
		a.MustEncodeProblem(w, req, err)
		return
	}

	a.MustEncodeJson(w, resp, http.StatusCreated)
}

func (a *App) GetRecords(w http.ResponseWriter, req *http.Request) {
	var (
		accountId uint64
		resp      svc.RecordsResponse
		err       error
	)

	if accountId, err = strconv.ParseUint(req.URL.Query().Get("accountId"), 10, 64); err != nil {
		a.MustEncodeProblem(w, req, ledger.NewErrorWithFields(
			ledger.ErrAccountValidation,
			"Invalid account Id provided",
			err, map[string]string{"accountId": req.URL.Query().Get("accountId")},
		))
		return
	}

	if req.URL.Query().Has("latest") {
		if resp, err = a.RecordService.GetRecords(req.Context(), ledger.AccountId(accountId)); err != nil {
			a.MustEncodeProblem(w, req, err)
			return
		}
	} else {
		log.Printf("Not implemented")
		// TODO search
	}

	a.MustEncodeJson(w, resp, http.StatusOK)
}
