package persistence

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
	dao "github.com/w-k-s/simple-budget-tracker/pkg/persistence"
)

type accountRecord struct {
	id         ledger.AccountId
	name       string
	currency   string
	createdBy  ledger.UserId
	createdAt  time.Time
	modifiedBy sql.NullInt64
	modifiedAt sql.NullTime
	version    ledger.Version
}

func (ar accountRecord) Id() ledger.AccountId {
	return ar.id
}

func (ar accountRecord) Name() string {
	return ar.name
}

func (ar accountRecord) Currency() string {
	return ar.currency
}

func (ar accountRecord) CreatedBy() ledger.UserId {
	return ar.createdBy
}

func (ar accountRecord) CreatedAtUTC() time.Time {
	return ar.createdAt
}

func (ar accountRecord) ModifiedBy() ledger.UserId {
	if ar.modifiedBy.Valid {
		return ledger.UserId(ar.modifiedBy.Int64)
	}
	return ledger.UserId(0)
}

func (ar accountRecord) ModifiedAtUTC() time.Time {
	if ar.modifiedAt.Valid {
		return ar.modifiedAt.Time
	}
	return time.Time{}
}

func (ar accountRecord) Version() ledger.Version {
	return ar.version
}

type DefaultAccountDao struct {
	*RootDao
}

func MustOpenAccountDao(driverName, dataSourceName string) dao.AccountDao {
	var db *sql.DB
	var err error
	if db, err = sql.Open(driverName, dataSourceName); err != nil {
		log.Fatalf("Failed to connect to data source: %q with driver driver: %q. Reason: %s", dataSourceName, driverName, err)
	}
	return &DefaultAccountDao{&RootDao{db}}
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
	epoch := time.Time{}
	_, err := tx.Exec(
		`INSERT INTO budget.account 
			(id, user_id, name, currency, created_by, created_at, last_modified_by, last_modified_at, version) 
			VALUES 
			($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		a.Id(),
		userId,
		a.Name(),
		a.Currency(),
		a.CreatedBy(),
		a.CreatedAtUTC(),
		sql.NullInt64{
			Int64: int64(a.ModifiedBy()),
			Valid: a.ModifiedBy() != 0,
		},
		sql.NullTime{
			Time:  a.ModifiedAtUTC(),
			Valid: epoch != a.ModifiedAtUTC(),
		},
		a.Version(),
	)
	if err != nil {
		log.Printf("Failed to save account %v. Reason: %s", a, err)
		if _, ok := d.isDuplicateKeyError(err); ok {
			return ledger.NewError(ledger.ErrAccountNameDuplicated, "Account name must be unique", err)
		}
		return ledger.NewError(ledger.ErrDatabaseState, "Failed to save account", err)
	}
	return nil
}

func (d *DefaultAccountDao) GetAccountById(queryId ledger.AccountId) (ledger.Account, error) {
	var ar accountRecord

	err := d.db.QueryRow(
		`SELECT 
			id, 
			name, 
			currency, 
			created_by, 
			created_at, 
			last_modified_by, 
			last_modified_at, 
			version 
		FROM 
			budget.account 
		WHERE id = $1`,
		queryId,
	).
		Scan(&ar.id, &ar.name, &ar.currency, &ar.createdBy, &ar.createdAt, &ar.modifiedBy, &ar.modifiedAt, &ar.version)
	if err != nil {
		if err == sql.ErrNoRows {
			return ledger.Account{}, ledger.NewError(ledger.ErrAccountNotFound, fmt.Sprintf("Account with id %d not found", queryId), err)
		}
		return ledger.Account{}, ledger.NewError(ledger.ErrDatabaseState, fmt.Sprintf("Account with id %d not found", queryId), err)
	}

	return ledger.NewAccountFromRecord(ar)
}

func (d *DefaultAccountDao) GetAccountsByUserId(queryId ledger.UserId) ([]ledger.Account, error) {

	rows, err := d.db.Query(
		`SELECT 
			a.id, 
			a.name, 
			a.currency, 
			a.created_by, 
			a.created_at, 
			a.last_modified_by,
			 a.last_modified_at, 
			 a.version 
		FROM budget.account a 
		INNER JOIN budget.user u 
			ON a.user_id = u.id 
		WHERE 
			u.id = $1 
		ORDER BY a.id`,
		queryId,
	)
	if err != nil {
		return nil, ledger.NewError(ledger.ErrDatabaseState, fmt.Sprintf("Accounts for user id %d not found", queryId), err)
	}
	defer rows.Close()

	entities := make([]ledger.Account, 0)
	for rows.Next() {
		var ar accountRecord

		if err := rows.Scan(&ar.id, &ar.name, &ar.currency, &ar.createdBy, &ar.createdAt, &ar.modifiedBy, &ar.modifiedAt, &ar.version); err != nil {
			log.Printf("Error processign accounts for user %d. Reason: %s", queryId, err)
			continue
		}

		var account ledger.Account
		if account, err = ledger.NewAccountFromRecord(ar); err != nil {
			log.Printf("Error loading account with id: %d,  name: %q, currency: %q from database. Reason: %s", ar.id, ar.name, ar.currency, err)
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
