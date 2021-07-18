package application

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type status bool

const (
	up   status = true
	down status = false
)

func (s status) String() string {
	if s == up {
		return "UP"
	}
	return "DOWN"
}

func (s status) MarshalJSON() ([]byte, error) {
	switch s {
	case up:
		fallthrough
	case down:
		return json.Marshal(s.String())
	default:
		return nil, fmt.Errorf("invalid status: %s", s)
	}
}

func (s *status) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	switch str {
	case "UP":
		*s = up
	case "DOWN":
		*s = down
	default:
		return fmt.Errorf("invalid status: %s", str)
	}
	return nil
}

func (s status) HttpCode() int {
	switch s {
	case up:
		return http.StatusOK
	default:
		return http.StatusInternalServerError
	}
}

type StatusReport map[string]status

func (report StatusReport) overallStatus() status {
	overall := up
	for _, status := range report {
		overall = overall && status
	}
	return overall
}

func (app *App) HealthHandler(w http.ResponseWriter, req *http.Request) {

	dbConfig := app.Config().Database()

	report := make(StatusReport)
	report["database"] = databaseStatusReport(&dbConfig)

	app.MustEncodeJson(w, report, report.overallStatus().HttpCode())
}

func databaseStatusReport(dbConfig *DBConfig) status {
	var (
		db  *sql.DB
		err error
	)

	if db, err = sql.Open(
		dbConfig.DriverName(),
		dbConfig.ConnectionString()); err != nil {
		log.Printf("Failed to connect to database for health check. Reason: %s", err)
		return down
	}

	db.SetMaxIdleConns(0) // Required, otherwise pinging will result in EOF
	if err = PingWithBackOff(db); err != nil {
		log.Printf("Ping failed for health check. Reason: %s", err)
		return down
	}

	return up
}
