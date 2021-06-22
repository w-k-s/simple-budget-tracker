package application

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/w-k-s/simple-budget-tracker/core"
)

type DefaultAccountDao struct {
	db *sql.DB
}

func MustOpenAccountDao(driverName, dataSourceName string) *DefaultAccountDao {
	var db *sql.DB
	var err error
	if db, err = sql.Open(driverName, dataSourceName); err != nil {
		log.Fatalf("Failed to connect to data source: %q with driver driver: %q. Reason: %s", dataSourceName, driverName, err)
	}
	return &DefaultAccountDao{db}
}

func (d DefaultAccountDao) Close() error {
	return d.db.Close()
}

func (d *DefaultAccountDao) NewAccountId() (core.AccountId, error) {
	var accountId core.AccountId
	err := d.db.QueryRow("SELECT nextval('budget.account_id')").Scan(&accountId)
	if err != nil {
		log.Printf("Failed to assign account id. Reason; %s", err)
		return 0, core.NewError(core.ErrDatabaseState, "Failed to assign account id", err)
	}
	return accountId, err
}

func (d *DefaultAccountDao) SaveTx(userId core.UserId, a *core.Account, tx *sql.Tx) error {
	_, err := tx.Exec("INSERT INTO budget.account (id, user_id, name, currency) VALUES ($1, $2, $3, $4)", a.Id(), userId, a.Name(), a.Currency())
	if err != nil {
		log.Printf("Failed to save account %v. Reason: %s", a, err)
		if _, ok := isDuplicateKeyError(err); ok {
			return core.NewError(core.ErrAccountNameDuplicated, "Account name must be unique", err)
		}
		return core.NewError(core.ErrDatabaseState, "Failed to save account", err)
	}
	return nil
}

func (d *DefaultAccountDao) GetAccountById(queryId core.AccountId) (*core.Account, error) {
	var accountId core.AccountId
	var name string
	var currency string

	err := d.db.QueryRow("SELECT id, name, currency FROM budget.account WHERE id = $1", queryId).Scan(&accountId, &name, &currency)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, core.NewError(core.ErrAccountNotFound, fmt.Sprintf("Account with id %d not found", queryId), err)
		}
		return nil, core.NewError(core.ErrDatabaseState, fmt.Sprintf("Account with id %d not found", queryId), err)
	}

	return core.NewAccount(accountId, name, currency)
}

func (d *DefaultAccountDao) GetAccountsByUserId(queryId core.UserId) ([]*core.Account, error) {

	rows, err := d.db.Query("SELECT a.id, a.name, a.currency FROM budget.account a INNER JOIN budget.user u ON a.user_id = u.id WHERE u.id = $1 ORDER BY a.id", queryId)
	if err != nil {
		return nil, core.NewError(core.ErrDatabaseState, fmt.Sprintf("Accounts for user id %d not found", queryId), err)
	}
	defer rows.Close()

	entities := make([]*core.Account, 0)
	for rows.Next() {
		var id core.AccountId
		var name string
		var currency string

		if err := rows.Scan(&id, &name, &currency); err != nil {
			log.Printf("Error processign accounts for user %d. Reason: %s", queryId, err)
			continue
		}

		var account *core.Account
		if account, err = core.NewAccount(id, name, currency); err != nil {
			log.Printf("Error loading account with id: %d,  name: %q, currency: %q from database. Reason: %s", id, name, currency, err)
			continue
		}

		entities = append(entities, account)
	}

	return entities, nil
}

func (d *DefaultAccountDao) Save(userId core.UserId, a *core.Account) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	err = d.SaveTx(userId, a, tx)
	if err == nil {
		err = tx.Commit()
	}
	return err
}
