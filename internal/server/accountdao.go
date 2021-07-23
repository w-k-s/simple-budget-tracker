package server

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
	dao "github.com/w-k-s/simple-budget-tracker/pkg/persistence"
)

type DefaultAccountDao struct {
	db *sql.DB
}

func MustOpenAccountDao(driverName, dataSourceName string) dao.AccountDao {
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

func (d *DefaultAccountDao) NewAccountId() (ledger.AccountId, error) {
	var accountId ledger.AccountId
	err := d.db.QueryRow("SELECT nextval('budget.account_id')").Scan(&accountId)
	if err != nil {
		log.Printf("Failed to assign account id. Reason; %s", err)
		return 0, ledger.NewError(ledger.ErrDatabaseState, "Failed to assign account id", err)
	}
	return accountId, err
}

func (d *DefaultAccountDao) SaveTx(userId ledger.UserId, a *ledger.Account, tx *sql.Tx) error {
	_, err := tx.Exec("INSERT INTO budget.account (id, user_id, name, currency) VALUES ($1, $2, $3, $4)", a.Id(), userId, a.Name(), a.Currency())
	if err != nil {
		log.Printf("Failed to save account %v. Reason: %s", a, err)
		if _, ok := isDuplicateKeyError(err); ok {
			return ledger.NewError(ledger.ErrAccountNameDuplicated, "Account name must be unique", err)
		}
		return ledger.NewError(ledger.ErrDatabaseState, "Failed to save account", err)
	}
	return nil
}

func (d *DefaultAccountDao) GetAccountById(queryId ledger.AccountId) (*ledger.Account, error) {
	var accountId ledger.AccountId
	var name string
	var currency string

	err := d.db.QueryRow("SELECT id, name, currency FROM budget.account WHERE id = $1", queryId).Scan(&accountId, &name, &currency)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ledger.NewError(ledger.ErrAccountNotFound, fmt.Sprintf("Account with id %d not found", queryId), err)
		}
		return nil, ledger.NewError(ledger.ErrDatabaseState, fmt.Sprintf("Account with id %d not found", queryId), err)
	}

	return ledger.NewAccount(accountId, name, currency)
}

func (d *DefaultAccountDao) GetAccountsByUserId(queryId ledger.UserId) ([]*ledger.Account, error) {

	rows, err := d.db.Query("SELECT a.id, a.name, a.currency FROM budget.account a INNER JOIN budget.user u ON a.user_id = u.id WHERE u.id = $1 ORDER BY a.id", queryId)
	if err != nil {
		return nil, ledger.NewError(ledger.ErrDatabaseState, fmt.Sprintf("Accounts for user id %d not found", queryId), err)
	}
	defer rows.Close()

	entities := make([]*ledger.Account, 0)
	for rows.Next() {
		var id ledger.AccountId
		var name string
		var currency string

		if err := rows.Scan(&id, &name, &currency); err != nil {
			log.Printf("Error processign accounts for user %d. Reason: %s", queryId, err)
			continue
		}

		var account *ledger.Account
		if account, err = ledger.NewAccount(id, name, currency); err != nil {
			log.Printf("Error loading account with id: %d,  name: %q, currency: %q from database. Reason: %s", id, name, currency, err)
			continue
		}

		entities = append(entities, account)
	}

	return entities, nil
}

func (d *DefaultAccountDao) Save(userId ledger.UserId, a *ledger.Account) error {
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
