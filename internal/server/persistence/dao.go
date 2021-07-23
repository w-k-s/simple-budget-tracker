package persistence

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/lib/pq"
	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
)

type RootDao struct {
	db *sql.DB
}

func (r *RootDao) BeginTx() (*sql.Tx, error) {
	var (
		tx  *sql.Tx
		err error
	)
	if tx, err = r.db.Begin(); err != nil {
		return nil, ledger.NewError(ledger.ErrDatabaseState, "Failed to begin transaction", err)
	}
	return tx, nil
}

func (r *RootDao) MustBeginTx() *sql.Tx {
	var (
		tx  *sql.Tx
		err error
	)

	if tx, err = r.db.Begin(); err != nil {
		log.Fatalf("Failed to begin transaction. Reason: %s", err)
	}
	return tx
}

func (r *RootDao) isDuplicateKeyError(err error) (string, bool) {
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
		return pqErr.Detail, true
	}
	return "", false
}

func PingWithBackOff(db *sql.DB) error {
	var ping backoff.Operation = func() error {
		err := db.Ping()
		if err != nil {
			log.Printf("DB is not ready...backing off...: %s", err)
			return err
		}
		log.Println("DB is ready!")
		return nil
	}

	exponentialBackoff := &backoff.ExponentialBackOff{
		InitialInterval:     backoff.DefaultInitialInterval,
		RandomizationFactor: backoff.DefaultRandomizationFactor,
		Multiplier:          backoff.DefaultMultiplier,
		MaxInterval:         time.Duration(50) * time.Millisecond,
		MaxElapsedTime:      time.Duration(300) * time.Millisecond,
		Clock:               backoff.SystemClock,
	}
	exponentialBackoff.Reset()

	var err error
	if err = backoff.Retry(ping, exponentialBackoff); err != nil {
		return fmt.Errorf("failed to connect to database after multiple retries. Reason: %w", err)
	}
	return nil
}
