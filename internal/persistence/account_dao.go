package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"
	"github.com/w-k-s/simple-budget-tracker/pkg"
	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
	dao "github.com/w-k-s/simple-budget-tracker/pkg/persistence"
)

type DefaultAccountDao struct {
	*RootDao
}

func MustOpenAccountDao(db *sql.DB) dao.AccountDao {
	return &DefaultAccountDao{&RootDao{db}}
}

func (d *DefaultAccountDao) NewAccountId(tx *sql.Tx) (ledger.AccountId, error) {
	var accountId ledger.AccountId
	err := d.db.QueryRow("SELECT nextval('budget.account_id')").Scan(&accountId)
	if err != nil {
		log.Printf("Failed to assign account id. Reason; %s", err)
		return 0, fmt.Errorf("Failed to assign account id: %w", err)
	}
	return accountId, err
}

func (d *DefaultAccountDao) SaveTx(ctx context.Context, userId ledger.UserId, a ledger.Accounts, tx *sql.Tx) error {

	stmt, err := tx.Prepare(pq.CopyInSchema(
		"budget",
		"account",
		"id",
		"user_id",
		"name",
		"account_type",
		"currency",
		"created_by",
		"created_at",
		"last_modified_by",
		"last_modified_at",
		"version",
	))
	if err != nil {
		return err
	}

	defer stmt.Close()

	epoch := time.Time{}
	for _, account := range a {
		_, err = stmt.Exec(
			account.Id(),
			userId,
			account.Name(),
			string(account.Type()),
			account.Currency(),
			account.CreatedBy().String(),
			account.CreatedAtUTC(),
			sql.NullString{
				String: account.ModifiedBy().String(),
				Valid:  account.ModifiedBy() != ledger.UpdatedBy{},
			},
			sql.NullTime{
				Time:  account.ModifiedAtUTC(),
				Valid: epoch != account.ModifiedAtUTC(),
			},
			account.Version(),
		)
		if err != nil {
			log.Printf("Failed to save category %q for user id %d. Reason: %q", account.Name(), userId, err)
		}

	}

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	return err
}

func (d *DefaultAccountDao) GetAccountsByUserId(ctx context.Context, queryId ledger.UserId, tx *sql.Tx) (ledger.Accounts, error) {

	rows, err := tx.QueryContext(
		ctx,
		`SELECT 
			a.id, 
			a.name, 
			a.account_type,
			a.currency, 
			(SELECT SUM(r.amount_minor_units) FROM budget.record r WHERE r.account_id = a.id),
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
		log.Printf("Error querying for accounts for user %d. Reason: %s", queryId, err)
		return nil, pkg.NewSystemError(pkg.ErrDatabaseState, fmt.Sprintf("Accounts for user id %d not found", queryId), err)
	}
	defer rows.Close()

	entities := make(ledger.Accounts, 0)
	for rows.Next() {
		var ar accountRecord

		if err := rows.Scan(&ar.id, &ar.name, &ar.accountType, &ar.currency, &ar.currentBalanceMinorUnits, &ar.createdBy, &ar.createdAt, &ar.modifiedBy, &ar.modifiedAt, &ar.version); err != nil {
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

func (d *DefaultAccountDao) GetAccountById(ctx context.Context, queryId ledger.AccountId, userId ledger.UserId, tx *sql.Tx) (ledger.Account, error) {
	var ar accountRecord
	err := tx.QueryRowContext(
		ctx,
		`SELECT 
			a.id, 
			a.name, 
			a.account_type,
			a.currency, 
			(SELECT SUM(r.amount_minor_units) FROM budget.record r WHERE r.account_id = a.id),
			a.created_by, 
			a.created_at, 
			a.last_modified_by,
			a.last_modified_at, 
			a.version 
		FROM budget.account a 
		LEFT JOIN budget.user u 
			ON a.user_id = u.id 
		LEFT JOIN budget.record r
			ON a.id = r.account_id
		WHERE 
			a.id = $1 
		AND a.user_id = $2
		ORDER BY a.id`,
		queryId, userId,
	).Scan(&ar.id, &ar.name, &ar.accountType, &ar.currency, &ar.currentBalanceMinorUnits, &ar.createdBy, &ar.createdAt, &ar.modifiedBy, &ar.modifiedAt, &ar.version)
	if err != nil {
		log.Printf("Failed to load account id %d for user %d. Reason: %s", queryId, userId, err)
		if err == sql.ErrNoRows {
			return ledger.Account{}, pkg.ValidationErrorWithError(pkg.ErrAccountNotFound, "Account not found", err)
		}
		return ledger.Account{}, pkg.NewSystemError(pkg.ErrDatabaseState, "Error loading account", err)
	}

	return ledger.NewAccountFromRecord(ar)
}

func (d *DefaultAccountDao) Save(ctx context.Context, userId ledger.UserId, as ledger.Accounts) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	err = d.SaveTx(ctx, userId, as, tx)
	if err == nil {
		err = tx.Commit()
	}
	return err
}

func (d *DefaultAccountDao) GetCurrenciesOfAccounts(
	ctx context.Context,
	accountIds ledger.AccountIds,
	userid ledger.UserId,
	tx *sql.Tx,
) (map[ledger.AccountId]ledger.Currency, error) {

	rows, err := tx.QueryContext(
		ctx,
		`SELECT 
			a.id, 
			a.currency 
		FROM budget.account a 
		INNER JOIN budget.user u 
			ON a.user_id = u.id 
		WHERE 
			a.id = ANY($1) 
		AND u.id = $2`,
		pq.Array(accountIds),
		userid,
	)
	if err != nil {
		log.Printf("Error querying for accounts for user %d. Reason: %s", userid, err)
		return nil, fmt.Errorf("Failed to query database: %w", err)
	}
	defer rows.Close()

	entities := make(map[ledger.AccountId]ledger.Currency, 0)
	for rows.Next() {
		accountId := ledger.AccountId(0)
		currency := ""

		if err := rows.Scan(
			&accountId,
			&currency,
		); err != nil {
			log.Printf("Error retrieving account id and currency for user %d. Reason: %s", userid, err)
			continue
		}

		entities[accountId] = ledger.MustMakeCurrency(ledger.MakeCurrency(currency))
	}

	return entities, nil
}
