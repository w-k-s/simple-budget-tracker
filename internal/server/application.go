package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
	cfg "github.com/w-k-s/simple-budget-tracker/internal/config"
	dao "github.com/w-k-s/simple-budget-tracker/internal/persistence"
	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
	svc "github.com/w-k-s/simple-budget-tracker/pkg/services"
	_ "github.com/w-k-s/simple-budget-tracker/statik"
	"schneider.vip/problem"
)

type App struct {
	config            *cfg.Config
	UserService       svc.UserService
	AccountService    svc.AccountService
	CategoriesService svc.CategoriesService
	RecordService     svc.RecordService
}

func (app *App) Config() *cfg.Config {
	return app.config
}

func Init(config *cfg.Config) (*App, error) {
	if config == nil {
		return nil, fmt.Errorf("configuration is required. Got %v", nil)
	}

	db, err := sql.Open(
		config.Database().DriverName(),
		config.Database().ConnectionString(),
	); 
	if err != nil {
		log.Fatalf("Failed to connect to data source: %q with driver driver: %q. Reason: %s", dataSourceName, driverName, err)
	}

	userDao := dao.MustOpenUserDao(db)
	userService, err := svc.NewUserService(userDao); 
	if err != nil {
		return nil, fmt.Errorf("failed to initiaise user service. Reason: %w", err)
	}

	accountDao := dao.MustOpenAccountDao(db)
	accountService, err := svc.NewAccountService(accountDao); 
	if err != nil {
		return nil, fmt.Errorf("failed to initiaise account service. Reason: %w", err)
	}

	categoryDao := dao.MustOpenCategoryDao(db)
	categoriesService, err := svc.NewCategoriesService(categoryDao); 
	if err != nil {
		return nil, fmt.Errorf("failed to initiaise categories service. Reason: %w", err)
	}

	recordDao := dao.MustOpenRecordDao(db)
	recordService, err := svc.NewRecordService(
		recordDao, 
		accountDao, 
		categoryDao, 
		config.Gpt().ApiKey(),
	); 
	if err != nil {
		return nil, fmt.Errorf("failed to initiaise record service. Reason: %w", err)
	}

	log.Printf("--- Application Initialized ---")
	return &App{
		config:            config,
		UserService:       userService,
		AccountService:    accountService,
		CategoriesService: categoriesService,
		RecordService:     recordService,
	}, nil
}

func (app *App) Router() *mux.Router {
	r := mux.NewRouter()

	r.Use(app.AuthenticationMiddleware)

	r.HandleFunc("/health", app.HealthHandler)

	users := r.PathPrefix("/api/v1/user").Subrouter()
	users.HandleFunc("", app.RegisterUser).
		Methods("POST")

	accounts := r.PathPrefix("/api/v1/accounts").Subrouter()
	accounts.HandleFunc("", app.RegisterAccounts).
		Methods("POST")
	accounts.HandleFunc("", app.GetAccounts).
		Methods("GET")

	categories := r.PathPrefix("/api/v1/categories").Subrouter()
	categories.HandleFunc("", app.CreateCategories).
		Methods("POST")
	categories.HandleFunc("", app.GetCategories).
		Methods("GET")

	records := r.PathPrefix("/api/v1/accounts/{accountId}/records").Subrouter()
	records.HandleFunc("", app.CreateRecord).
		Methods("POST")
	records.HandleFunc("/gpt", app.CreateRecordRequestWithChatGPT).
		Methods("POST")
	records.HandleFunc("", app.GetRecords).
		Methods("GET")

	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}
	staticServer := http.FileServer(statikFS)
	sh := http.StripPrefix("/swaggerui/", staticServer)
	r.PathPrefix("/swaggerui/").Handler(sh)

	return r
}

func (app *App) MustEncodeJson(w http.ResponseWriter, v interface{}, status int) {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(true)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := encoder.Encode(v); err != nil {
		log.Fatalf("Failed to encode json '%v'. Reason: %s", v, err)
	}
}

func (app *App) DecodeJsonOrSendBadRequest(w http.ResponseWriter, req *http.Request, v interface{}) bool {
	decoder := json.NewDecoder(req.Body)
	decoder.UseNumber()
	if err := decoder.Decode(v); err != nil {
		app.MustEncodeProblem(w, req, ledger.NewError(ledger.ErrRequestUnmarshallingFailed, "Failed to parse request", err))
		return false
	}
	return true
}

func (app *App) MustEncodeProblem(w http.ResponseWriter, req *http.Request, err error) {

	log.Printf("Error: %s", err.Error())

	title := ledger.ErrUnknown.Name()
	code := ledger.ErrUnknown
	detail := err.Error()
	opts := []problem.Option{}
	status := ledger.ErrUnknown.Status()

	if coreError, ok := err.(ledger.Error); ok {
		title = coreError.Code().Name()
		code = coreError.Code()
		detail = coreError.Error()
		status = coreError.Code().Status()

		for key, value := range coreError.Fields() {
			opts = append(opts, problem.Custom(key, value))
		}
	}

	if _, problemError := problem.New(
		problem.Type(fmt.Sprintf("/api/v1/problems/%d", code)),
		problem.Status(status),
		problem.Instance(req.URL.Path),
		problem.Title(title),
		problem.Detail(detail),
	).
		Append(opts...).
		WriteTo(w); problemError != nil {
		log.Printf("Failed to encode problem '%v'. Reason: %s", err, problemError)
	}
}

func (a *App) AuthenticationMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if authorization := r.Header.Get("Authorization"); len(authorization) != 0 {
			userId, err := strconv.ParseUint(authorization, 10, 64)
			if err != nil {
				log.Printf("Failed to parse userId %q in authorization header", authorization)
			}
			r = r.WithContext(context.WithValue(r.Context(), svc.CtxUserId, ledger.UserId(userId)))
		}

		h.ServeHTTP(w, r)
	})
}
