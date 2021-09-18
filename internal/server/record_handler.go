package server

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
	svc "github.com/w-k-s/simple-budget-tracker/pkg/services"
)

func (a *App) CreateRecord(w http.ResponseWriter, req *http.Request) {

	var (
		accountId           ledger.AccountId
		createRecordRequest svc.CreateRecordRequest
		resp                svc.RecordResponse
		err                 error
		ok                  bool
	)

	if accountId, ok = a.getAccountIdOrBadRequest(w, req); !ok {
		return
	}

	if ok = a.DecodeJsonOrSendBadRequest(w, req, &createRecordRequest); !ok {
		return
	}

	req = req.WithContext(svc.SetAccountId(req.Context(), accountId))
	if resp, err = a.RecordService.CreateRecord(req.Context(), createRecordRequest); err != nil {
		a.MustEncodeProblem(w, req, err)
		return
	}

	a.MustEncodeJson(w, resp, http.StatusCreated)
}

func (a *App) GetRecords(w http.ResponseWriter, req *http.Request) {
	var (
		accountId ledger.AccountId
		resp      svc.RecordsResponse
		err       error
		ok        bool
	)

	if accountId, ok = a.getAccountIdOrBadRequest(w, req); !ok {
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

func (a *App) getAccountIdOrBadRequest(w http.ResponseWriter, req *http.Request) (ledger.AccountId, bool) {
	var (
		accountId uint64
		err       error
	)

	params := mux.Vars(req)
	if accountId, err = strconv.ParseUint(params["accountId"], 10, 64); err != nil {
		a.MustEncodeProblem(w, req, ledger.NewErrorWithFields(
			ledger.ErrServiceAccountIdRequired,
			"Invalid or no account Id provided",
			err, map[string]string{"accountId": req.URL.Query().Get("accountId")},
		))
		return 0, false
	}
	return ledger.AccountId(accountId), true
}
