package application

import (
	"database/sql"
	"log"

	"github.com/w-k-s/simple-budget-tracker/core"
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
		return nil, core.NewError(core.ErrDatabaseState, "Failed to begin transaction", err)
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
