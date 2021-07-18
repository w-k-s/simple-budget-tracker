package application

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type App struct {
	config *Config
}

func (app *App) Config() *Config {
	return app.config
}

func Init(config *Config) (*App, error) {
	if config == nil {
		return nil, fmt.Errorf("configuration is required. Got %v", nil)
	}

	log.Printf("--- Application Initialized ---")
	return &App{
		config: config,
	}, nil
}

func (app *App) Router() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/health", app.HealthHandler)

	return r
}

func (app *App) MustEncodeJson(w http.ResponseWriter, v interface{}, status int) {
	encoder := json.NewEncoder(w)
	w.WriteHeader(status)
	if err := encoder.Encode(v); err != nil {
		log.Fatalf("Failed to encode json '%v'. Reason: %s", v, err)
	}
}
