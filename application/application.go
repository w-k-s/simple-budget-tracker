package application

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/w-k-s/simple-budget-tracker/core"
	"schneider.vip/problem"
)

type App struct {
	config      *Config
	UserService core.UserService
}

func (app *App) Config() *Config {
	return app.config
}

func Init(config *Config) (*App, error) {
	if config == nil {
		return nil, fmt.Errorf("configuration is required. Got %v", nil)
	}

	var (
		userService core.UserService
		err         error
	)

	if userService, err = core.NewUserService(MustOpenUserDao(
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

	// accounts := r.PathPrefix("/api/v1/accounts").Subrouter()
	// accounts.HandleFunc("/", app.CreateAccount).
	// 	Methods("POST")

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

	title := core.ErrUnknown.Name()
	code := core.ErrUnknown
	detail := err.Error()
	opts := []problem.Option{}

	if coreError, ok := err.(core.Error); ok {
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
