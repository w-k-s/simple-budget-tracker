package persistence

import (
	"database/sql"
	"log"

	dao "github.com/w-k-s/simple-budget-tracker/pkg/persistence"
)

type defaultDaoFactory struct {
	db *sql.DB
}

func NewFactory(db *sql.DB) dao.Factory {
	if db == nil {
		panic("Cannot instantiate DaoFactory, db is nil")
	}

	return &defaultDaoFactory{
		db: db,
	}
}

func (d *defaultDaoFactory) BeginTx() (*sql.Tx, error) {
	return d.db.Begin()
}

func (d *defaultDaoFactory) GetBudgetDao(tx *sql.Tx) dao.BudgetDao {
	if tx == nil {
		tx, err := d.db.Begin()
		if err != nil {
			log.Fatalf("Cannot begin transaction when creating budget dao. Reason: %q", err)
		}
		return &defaultBudgetDao{tx}
	}
	return &defaultBudgetDao{tx}
}
