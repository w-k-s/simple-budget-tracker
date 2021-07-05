package application

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/w-k-s/simple-budget-tracker/core"
)

type DefaultRecordDao struct {
	db *sql.DB
}

func MustOpenRecordDao(driverName, dataSourceName string) core.RecordDao {
	var db *sql.DB
	var err error
	if db, err = sql.Open(driverName, dataSourceName); err != nil {
		log.Fatalf("Failed to connect to data source: %q with driver driver: %q. Reason: %s", dataSourceName, driverName, err)
	}
	return &DefaultRecordDao{db}
}

func (d DefaultRecordDao) Close() error {
	return d.db.Close()
}

func (d *DefaultRecordDao) NewRecordId() (core.RecordId, error) {
	var recordId core.RecordId
	err := d.db.QueryRow("SELECT nextval('budget.record_id')").Scan(&recordId)
	if err != nil {
		log.Printf("Failed to assign record id. Reason; %s", err)
		return 0, core.NewError(core.ErrDatabaseState, "Failed to assign record id", err)
	}
	return recordId, err
}

func (d *DefaultRecordDao) Save(accountId core.AccountId, r *core.Record) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	err = d.SaveTx(accountId, r, tx)
	if err == nil {
		err = tx.Commit()
	}
	return err
}

func (d *DefaultRecordDao) SaveTx(accountId core.AccountId, r *core.Record, tx *sql.Tx) error {
	amountMinorUnits, _ := r.Amount().MinorUnits()
	_, err := tx.Exec("INSERT INTO budget.record (id, account_id, category_id, note, currency, amount_minor_units, date, type, beneficiary_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
		r.Id(),
		accountId,
		r.Category().Id(),
		r.Note(),
		r.Amount().Currency().CurrencyCode(),
		amountMinorUnits,
		r.DateUTC(),
		r.Type(),
		sql.NullInt64{
			Int64: int64(r.BeneficiaryId()),
			Valid: r.BeneficiaryId() != 0,
		},
	)
	if err != nil {
		log.Printf("Failed to save record %v. Reason: %s", r, err)
		return core.NewError(core.ErrDatabaseState, "Failed to save record", err)
	}
	return nil
}

func (d *DefaultRecordDao) Search(id core.AccountId, search core.RecordSearch) (core.Records, error) {
	// TODO
	return nil, fmt.Errorf("To implement")
}

func (d *DefaultRecordDao) GetRecordsForLastPeriod(accountId core.AccountId) (core.Records, error) {
	// TODO
	return nil, fmt.Errorf("To implement")
}

func (d *DefaultRecordDao) GetRecordsForMonth(queryId core.AccountId, month int, year int) (core.Records, error) {
	fromDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	toDate := fromDate.AddDate(0, 1, -1)
	rows, err := d.db.Query("SELECT r.id, r.category_id, c.name, r.note, r.currency, r.amount_minor_units, r.date, r.type, r.beneficiary_id FROM budget.record r LEFT JOIN budget.account a ON r.account_id = a.id LEFT JOIN budget.category c ON r.category_id = c.id WHERE a.id = $1 AND r.date >= $2 AND r.date <= $3 ORDER BY r.date DESC",
		queryId,
		fromDate.Format("2006-01-02"),
		toDate.Format("2006-01-02"),
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return core.Records{}, nil
		}
		return nil, core.NewError(core.ErrDatabaseState, fmt.Sprintf("Records for account id %d not found", queryId), err)
	}
	defer rows.Close()

	entities := make([]*core.Record, 0)
	for rows.Next() {

		var (
			recordId            core.RecordId
			categoryId          core.CategoryId
			categoryName        string
			note                string
			currency            string
			amountMinorUnits    int64
			date                time.Time
			recordType          core.RecordType
			beneficiaryIdOrNull sql.NullInt64
			beneficiaryId       core.AccountId
		)

		if err = rows.Scan(&recordId, &categoryId, &categoryName, &note, &currency, &amountMinorUnits, &date, &recordType, &beneficiaryIdOrNull); err != nil {
			log.Printf("Error processing records for account %d. Reason: %s", queryId, err)
			continue
		}

		if beneficiaryIdOrNull.Valid {
			beneficiaryId = core.AccountId(beneficiaryIdOrNull.Int64)
		}

		var (
			amount   core.Money
			category *core.Category
			record   *core.Record
		)

		if amount, err = core.NewMoney(currency, amountMinorUnits); err != nil {
			log.Printf("Error loading record id: %d for account id %d, currency: %q, amount (minor units): %d from database. Reason: %s", recordId, queryId, currency, amountMinorUnits, err)
			continue
		}

		if category, err = core.NewCategory(categoryId, categoryName); err != nil {
			log.Printf("Error loading record id: %d for account id %d, category id: %d, category name: %s from database. Reason: %s", recordId, queryId, categoryId, categoryName, err)
			continue
		}

		if record, err = core.NewRecord(recordId, note, category, amount, date, recordType, beneficiaryId); err != nil {
			log.Printf("Error loading account with id: %d from database. Reason: %s", recordId, err)
			continue
		}

		entities = append(entities, record)
	}

	return entities, nil
}
