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
	"github.com/w-k-s/simple-budget-tracker/pkg"
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

	err := cfg.ConfigureLogging()
	if err != nil {
		log.Fatalf("failed to configure logging. Reason: %s", err)
	}

	db, err := sql.Open(
		config.Database().DriverName(),
		config.Database().ConnectionString(),
	)
	if err != nil {
		log.Fatalf("Failed to connect to data source: %q with driver driver: %q. Reason: %s", config.Database().ConnectionString(), config.Database().DriverName(), err)
	}
	dao.MustRunMigrations(db, config.Database())

	userDao := dao.MustOpenUserDao(db)
	userService, err := svc.NewUserService(userDao)
	if err != nil {
		return nil, fmt.Errorf("failed to initiaise user service. Reason: %w", err)
	}

	accountDao := dao.MustOpenAccountDao(db)
	accountService, err := svc.NewAccountService(accountDao)
	if err != nil {
		return nil, fmt.Errorf("failed to initiaise account service. Reason: %w", err)
	}

	categoryDao := dao.MustOpenCategoryDao(db)
	categoriesService, err := svc.NewCategoriesService(categoryDao)
	if err != nil {
		return nil, fmt.Errorf("failed to initiaise categories service. Reason: %w", err)
	}

	recordDao := dao.MustOpenRecordDao(db)
	recordService, err := svc.NewRecordService(
		recordDao,
		accountDao,
		categoryDao,
		config.Gpt().ApiKey(),
	)
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
		app.MustEncodeProblem(w, req, pkg.NewSystemError(pkg.ErrRequestUnmarshallingFailed, "Failed to parse request", err))
		return false
	}
	return true
}

func (app *App) MustEncodeProblem(w http.ResponseWriter, req *http.Request, err error) {

	log.Printf("Error: %s", err.Error())

	opts := []problem.Option{}
	for key, value := range errorFields(err) {
		opts = append(opts, problem.Custom(key, value))
	}

	p := problem.New(
		problem.Type(fmt.Sprintf("/api/v1/problems/%d", errorCode(err))),
		problem.Status(errorStatus(err)),
		problem.Instance(req.URL.Path),
		problem.Title(errorTitle(err)),
		problem.Detail(errorDetail(err)),
	)
	p.Append(opts...)
	if _, encodingError := p.WriteTo(w); encodingError != nil {
		log.Printf("Failed to encode problem '%v'. Reason: %s", err, encodingError)
	}
}

func errorTitle(err error) string {
	if errorWithTitle, ok := err.(interface {
		Title() string
	}); ok {
		return errorWithTitle.Title()
	}
	return ""
}

func errorCode(err error) uint64 {
	if errorWithCode, ok := err.(interface {
		Code() uint64
	}); ok {
		return errorWithCode.Code()
	}
	return 0
}

func errorStatus(err error) int {
	if errorWithStatus, ok := err.(interface {
		StatusCode() int
	}); ok {
		return errorWithStatus.StatusCode()
	}
	return 500
}

func errorDetail(err error) string {
	if errorWithDetails, ok := err.(interface {
		Detail() string
	}); ok {
		return errorWithDetails.Detail()
	}
	return ""
}

func errorFields(err error) map[string]string {
	if errWithFields, ok := err.(interface {
		InvalidFields() map[string]string
	}); ok {
		return errWithFields.InvalidFields()
	}
	return map[string]string{}
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
