package server

import (
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
	dao "github.com/w-k-s/simple-budget-tracker/pkg/persistence"
)

type DefaultRecordDao struct {
	db *sql.DB
}

func MustOpenRecordDao(driverName, dataSourceName string) dao.RecordDao {
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

func (d *DefaultRecordDao) NewRecordId() (ledger.RecordId, error) {
	var recordId ledger.RecordId
	err := d.db.QueryRow("SELECT nextval('budget.record_id')").Scan(&recordId)
	if err != nil {
		log.Printf("Failed to assign record id. Reason; %s", err)
		return 0, ledger.NewError(ledger.ErrDatabaseState, "Failed to assign record id", err)
	}
	return recordId, err
}

func (d *DefaultRecordDao) Save(accountId ledger.AccountId, r *ledger.Record) error {
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

func (d *DefaultRecordDao) SaveTx(accountId ledger.AccountId, r *ledger.Record, tx *sql.Tx) error {
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
		return ledger.NewError(ledger.ErrDatabaseState, "Failed to save record", err)
	}
	return nil
}

func (d *DefaultRecordDao) Search(accountId ledger.AccountId, search dao.RecordSearch) (ledger.Records, error) {
	checkError := func(err error) bool {
		if err != nil {
			log.Printf("Error searching for records with criteria %v account id: %d. Reason: %s", search, accountId, err)
			return true
		}
		return false
	}

	if search.FromDate == nil {
		defaultFromDate := startOfCurrentMonth()
		search.FromDate = &defaultFromDate
	}

	if search.ToDate == nil {
		defaultToDate := time.Now().UTC()
		search.ToDate = &defaultToDate
	}

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	query := psql.Select(
		"r.id",
		"r.category_id",
		"c.name",
		"r.note",
		"r.currency",
		"r.amount_minor_units",
		"r.date",
		"r.type",
		"r.beneficiary_id",
	).
		From("budget.record r").
		LeftJoin("budget.category c ON c.id = r.category_id").
		Where(sq.Eq{
			"r.account_id": accountId,
		})

	if len(search.CategoryNames) != 0 {
		query = query.Where(
			sq.Eq{"c.name": search.CategoryNames},
		)
	}

	if len(search.RecordTypes) != 0 {
		query = query.Where(sq.Eq{"r.type": search.RecordTypes})
	}

	if len(search.SearchTerm) != 0 {
		keywords := strings.Split(search.SearchTerm, " ")

		likes := make([]sq.Sqlizer, 0, len(keywords))
		for _, keyword := range keywords {
			if len(keyword) == 0 {
				continue
			}
			if strings.Contains(keyword, "\"") || strings.Contains(keyword, ";") {
				// Crappy check against sql injection
				continue
			}
			likes = append(likes, sq.Like{"r.note": fmt.Sprintf("%%%s%%", keyword)})
		}

		query = query.Where(sq.Or(likes))
	}

	var (
		rows *sql.Rows
		err  error
	)

	if rows, err = query.OrderBy("r.date DESC").
		RunWith(d.db).
		Query(); checkError(err) {
		return ledger.Records{}, nil
	}

	defer rows.Close()

	entities := make([]*ledger.Record, 0)
	for rows.Next() {

		var (
			recordId            ledger.RecordId
			categoryId          ledger.CategoryId
			categoryName        string
			note                string
			currency            string
			amountMinorUnits    int64
			date                time.Time
			recordType          ledger.RecordType
			beneficiaryIdOrNull sql.NullInt64
			beneficiaryId       ledger.AccountId
		)

		if err = rows.Scan(&recordId, &categoryId, &categoryName, &note, &currency, &amountMinorUnits, &date, &recordType, &beneficiaryIdOrNull); err != nil {
			log.Printf("Error processing records for account %d. Reason: %s", accountId, err)
			continue
		}

		if beneficiaryIdOrNull.Valid {
			beneficiaryId = ledger.AccountId(beneficiaryIdOrNull.Int64)
		}

		var (
			amount   ledger.Money
			category *ledger.Category
			record   *ledger.Record
		)

		if amount, err = ledger.NewMoney(currency, amountMinorUnits); err != nil {
			log.Printf("Error loading record id: %d for account id %d, currency: %q, amount (minor units): %d from database. Reason: %s", recordId, accountId, currency, amountMinorUnits, err)
			continue
		}

		if category, err = ledger.NewCategory(categoryId, categoryName); err != nil {
			log.Printf("Error loading record id: %d for account id %d, category id: %d, category name: %s from database. Reason: %s", recordId, accountId, categoryId, categoryName, err)
			continue
		}

		if record, err = ledger.NewRecord(recordId, note, category, amount, date, recordType, beneficiaryId); err != nil {
			log.Printf("Error loading account with id: %d from database. Reason: %s", recordId, err)
			continue
		}

		entities = append(entities, record)
	}

	records := ledger.Records(entities)
	sort.Sort(records)

	return records, nil
}

func (d *DefaultRecordDao) GetRecordsForLastPeriod(accountId ledger.AccountId) (ledger.Records, error) {
	checkError := func(err error) bool {
		if err != nil {
			log.Printf("Error loading records for last period for account id: %d. Reason: %s", accountId, err)
			return true
		}
		return false
	}

	var rows *sql.Rows
	var err error
	if rows, err = d.db.Query("SELECT MAX(r.date) FROM budget.record r WHERE r.account_id = $1", accountId); checkError(err) {
		return ledger.Records{}, nil
	}

	defer rows.Close()
	rows.Next()

	var max time.Time
	if err = rows.Scan(&max); checkError(err) {
		return ledger.Records{}, nil
	}

	return d.GetRecordsForMonth(accountId, ledger.MakeCalendarMonthFromDate(max))
}

func (d *DefaultRecordDao) GetRecordsForMonth(queryId ledger.AccountId, month ledger.CalendarMonth) (ledger.Records, error) {
	fromDate := month.FirstDay()
	toDate := month.LastDay()
	rows, err := d.db.Query("SELECT r.id, r.category_id, c.name, r.note, r.currency, r.amount_minor_units, r.date, r.type, r.beneficiary_id FROM budget.record r LEFT JOIN budget.account a ON r.account_id = a.id LEFT JOIN budget.category c ON r.category_id = c.id WHERE a.id = $1 AND r.date >= $2 AND r.date <= $3 ORDER BY r.date DESC",
		queryId,
		fromDate.Format("2006-01-02"),
		toDate.Format("2006-01-02"),
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return ledger.Records{}, nil
		}
		return nil, ledger.NewError(ledger.ErrDatabaseState, fmt.Sprintf("Records for account id %d not found", queryId), err)
	}
	defer rows.Close()

	entities := make([]*ledger.Record, 0)
	for rows.Next() {

		var (
			recordId            ledger.RecordId
			categoryId          ledger.CategoryId
			categoryName        string
			note                string
			currency            string
			amountMinorUnits    int64
			date                time.Time
			recordType          ledger.RecordType
			beneficiaryIdOrNull sql.NullInt64
			beneficiaryId       ledger.AccountId
		)

		if err = rows.Scan(&recordId, &categoryId, &categoryName, &note, &currency, &amountMinorUnits, &date, &recordType, &beneficiaryIdOrNull); err != nil {
			log.Printf("Error processing records for account %d. Reason: %s", queryId, err)
			continue
		}

		if beneficiaryIdOrNull.Valid {
			beneficiaryId = ledger.AccountId(beneficiaryIdOrNull.Int64)
		}

		var (
			amount   ledger.Money
			category *ledger.Category
			record   *ledger.Record
		)

		if amount, err = ledger.NewMoney(currency, amountMinorUnits); err != nil {
			log.Printf("Error loading record id: %d for account id %d, currency: %q, amount (minor units): %d from database. Reason: %s", recordId, queryId, currency, amountMinorUnits, err)
			continue
		}

		if category, err = ledger.NewCategory(categoryId, categoryName); err != nil {
			log.Printf("Error loading record id: %d for account id %d, category id: %d, category name: %s from database. Reason: %s", recordId, queryId, categoryId, categoryName, err)
			continue
		}

		if record, err = ledger.NewRecord(recordId, note, category, amount, date, recordType, beneficiaryId); err != nil {
			log.Printf("Error loading account with id: %d from database. Reason: %s", recordId, err)
			continue
		}

		entities = append(entities, record)
	}

	records := ledger.Records(entities)
	sort.Sort(records)

	return records, nil
}

func startOfCurrentMonth() time.Time {
	return time.Now().UTC().AddDate(0, -1, time.Now().UTC().Day()+1)
}
