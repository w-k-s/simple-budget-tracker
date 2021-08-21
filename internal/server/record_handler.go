package server

import (
	"net/http"

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
		resp svc.CalendarMonthRecordsResponse
		err  error
	)

	if resp, err = a.RecordService.GetRecords(req.Context()); err != nil {
		a.MustEncodeProblem(w, req, err)
		return
	}

	a.MustEncodeJson(w, resp, http.StatusOK)
}
