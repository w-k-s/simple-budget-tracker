package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
	cfg "github.com/w-k-s/simple-budget-tracker/internal/config"
	dao "github.com/w-k-s/simple-budget-tracker/internal/server/persistence"
	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
	svc "github.com/w-k-s/simple-budget-tracker/pkg/services"
	_ "github.com/w-k-s/simple-budget-tracker/statik"
	"schneider.vip/problem"
)

type App struct {
	config      *cfg.Config
	UserService svc.UserService
}

func (app *App) Config() *cfg.Config {
	return app.config
}

func Init(config *cfg.Config) (*App, error) {
	if config == nil {
		return nil, fmt.Errorf("configuration is required. Got %v", nil)
	}

	var (
		userService svc.UserService
		err         error
	)

	if userService, err = svc.NewUserService(dao.MustOpenUserDao(
		config.Database().DriverName(),
		config.Database().ConnectionString(),
	)); err != nil {
		return nil, fmt.Errorf("failed to initiaise user service. Reason: %w", err)
	}

	log.Printf("--- Application Initialized ---")
	return &App{
		config:      config,
		UserService: userService,
	}, nil
}

func (app *App) Router() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/health", app.HealthHandler)

	users := r.PathPrefix("/api/v1/user").Subrouter()
	users.HandleFunc("", app.RegisterUser).
		Methods("POST")

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
	w.WriteHeader(status)
	if err := encoder.Encode(v); err != nil {
		log.Fatalf("Failed to encode json '%v'. Reason: %s", v, err)
	}
}

func (app *App) MustDecodeJson(r io.Reader, v interface{}) error {
	decoder := json.NewDecoder(r)
	return decoder.Decode(v)
}

func (app *App) MustEncodeProblem(w http.ResponseWriter, req *http.Request, err error, status int) {

	log.Printf("Error: %s", err.Error())

	title := ledger.ErrUnknown.Name()
	code := ledger.ErrUnknown
	detail := err.Error()
	opts := []problem.Option{}

	if coreError, ok := err.(ledger.Error); ok {
		title = coreError.Code().Name()
		code = coreError.Code()
		detail = coreError.Error()

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
