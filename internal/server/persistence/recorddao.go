package persistence

import (
	"context"
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

type recordRecord struct {
	id                ledger.RecordId
	note              string
	category          categoryRecord
	currency          string
	amountMinorUnits  int64
	date              time.Time
	recordType        ledger.RecordType
	sourceAccountId   sql.NullInt64
	beneficiaryId     sql.NullInt64
	transferReference sql.NullString
	createdBy         string
	createdAt         time.Time
	modifiedBy        sql.NullString
	modifiedAt        sql.NullTime
	version           ledger.Version
}

func (rr recordRecord) Id() ledger.RecordId {
	return rr.id
}

func (rr recordRecord) Note() string {
	return rr.note
}

func (rr recordRecord) Category() ledger.Category {
	var (
		category ledger.Category
		err      error
	)
	if category, err = ledger.NewCategoryFromRecord(rr.category); err != nil {
		log.Fatalf("Failed to parse category from database for record id %d. Reason: %s", rr.id, err)
	}
	return category
}

func (rr recordRecord) Amount() ledger.Money {
	var (
		amount ledger.Money
		err    error
	)
	if amount, err = ledger.NewMoney(rr.currency, rr.amountMinorUnits); err != nil {
		log.Fatalf("Failed to parse amount from database for record id %d. Reason: %s", rr.id, err)
	}
	return amount
}

func (rr recordRecord) DateUTC() time.Time {
	return rr.date
}

func (rr recordRecord) RecordType() ledger.RecordType {
	return rr.recordType
}

func (rr recordRecord) SourceAccountId() ledger.AccountId {
	if rr.sourceAccountId.Valid {
		return ledger.AccountId(rr.sourceAccountId.Int64)
	}
	return ledger.NoSourceAccount
}

func (rr recordRecord) BeneficiaryId() ledger.AccountId {
	if rr.beneficiaryId.Valid {
		return ledger.AccountId(rr.beneficiaryId.Int64)
	}
	return ledger.NoBeneficiaryAccount
}

func (rr recordRecord) TransferReference() ledger.TransferReference {
	if rr.beneficiaryId.Valid {
		return ledger.TransferReference(rr.transferReference.String)
	}
	return ledger.NoTransferReference
}

func (rr recordRecord) CreatedBy() ledger.UpdatedBy {
	var (
		updatedBy ledger.UpdatedBy
		err       error
	)
	if updatedBy, err = ledger.ParseUpdatedBy(rr.createdBy); err != nil {
		log.Fatalf("Invalid createdBy persisted for record %d: %s", rr.id, rr.createdBy)
	}
	return updatedBy
}

func (rr recordRecord) CreatedAtUTC() time.Time {
	return rr.createdAt
}

func (rr recordRecord) ModifiedBy() ledger.UpdatedBy {
	if !rr.modifiedBy.Valid {
		return ledger.UpdatedBy{}
	}
	var (
		updatedBy ledger.UpdatedBy
		err       error
	)
	if updatedBy, err = ledger.ParseUpdatedBy(rr.modifiedBy.String); err != nil {
		log.Fatalf("Invalid modifiedBy persisted for record %d: %s", rr.id, rr.ModifiedBy())
	}
	return updatedBy
}

func (rr recordRecord) ModifiedAtUTC() time.Time {
	if rr.modifiedAt.Valid {
		return rr.modifiedAt.Time
	}
	return time.Time{}
}

func (rr recordRecord) Version() ledger.Version {
	return rr.version
}

type DefaultRecordDao struct {
	RootDao
}

func MustOpenRecordDao(driverName, dataSourceName string) dao.RecordDao {
	var db *sql.DB
	var err error
	if db, err = sql.Open(driverName, dataSourceName); err != nil {
		log.Fatalf("Failed to connect to data source: %q with driver driver: %q. Reason: %s", dataSourceName, driverName, err)
	}
	return &DefaultRecordDao{RootDao{db}}
}

func (d DefaultRecordDao) Close() error {
	return d.db.Close()
}

func (d *DefaultRecordDao) NewRecordId(tx *sql.Tx) (ledger.RecordId, error) {
	var recordId ledger.RecordId
	err := tx.QueryRow("SELECT nextval('budget.record_id')").Scan(&recordId)
	if err != nil {
		log.Printf("Failed to assign record id. Reason; %s", err)
		return 0, ledger.NewError(ledger.ErrDatabaseState, "Failed to assign record id", err)
	}
	return recordId, err
}

func (d *DefaultRecordDao) Save(ctx context.Context, accountId ledger.AccountId, r ledger.Record) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	err = d.SaveTx(ctx, accountId, r, tx)
	if err == nil {
		err = tx.Commit()
	}
	return err
}

func (d *DefaultRecordDao) SaveTx(ctx context.Context, accountId ledger.AccountId, r ledger.Record, tx *sql.Tx) error {
	epoch := time.Time{}
	amountMinorUnits, _ := r.Amount().MinorUnits()
	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO budget.record 
		(id, account_id, category_id, note, currency, amount_minor_units, date, type, source_account_id, beneficiary_id, transfer_reference, created_by, created_at, last_modified_by, last_modified_at, version) 
		VALUES 
		($1, $2, $3, $4, (SELECT currency FROM budget.account WHERE id = $5), $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`,
		r.Id(),
		accountId,
		r.Category().Id(),
		r.Note(),
		accountId,
		amountMinorUnits,
		r.DateUTC(),
		r.Type(),
		sql.NullInt64{
			Int64: int64(r.SourceAccountId()),
			Valid: r.SourceAccountId() != 0,
		},
		sql.NullInt64{
			Int64: int64(r.BeneficiaryId()),
			Valid: r.BeneficiaryId() != 0,
		},
		sql.NullString{
			String: string(r.TransferReference()),
			Valid:  len(r.TransferReference()) != 0,
		},
		r.CreatedBy().String(),
		r.CreatedAtUTC(),
		sql.NullString{
			String: r.ModifiedBy().String(),
			Valid:  r.ModifiedBy() != ledger.UpdatedBy{},
		},
		sql.NullTime{
			Time:  r.ModifiedAtUTC(),
			Valid: epoch != r.ModifiedAtUTC(),
		},
		r.Version(),
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
		defaultFromDate := ledger.CurrentCalendarMonth().FirstDay()
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
		"c.created_by",
		"c.created_at",
		"c.last_modified_by",
		"c.last_modified_at",
		"c.version",
		"r.note",
		"r.currency",
		"r.amount_minor_units",
		"r.date",
		"r.type",
		"r.source_account_id",
		"r.beneficiary_id",
		"r.transfer_reference",
		"r.created_by",
		"r.created_at",
		"r.last_modified_by",
		"r.last_modified_at",
		"r.version",
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

	entities := make([]ledger.Record, 0)
	for rows.Next() {

		var (
			rr     recordRecord
			record ledger.Record
		)
		if err = rows.Scan(
			&rr.id,
			&rr.category.id,
			&rr.category.name,
			&rr.category.createdBy,
			&rr.category.createdAt,
			&rr.category.modifiedBy,
			&rr.category.modifiedAt,
			&rr.category.version,
			&rr.note,
			&rr.currency,
			&rr.amountMinorUnits,
			&rr.date,
			&rr.recordType,
			&rr.sourceAccountId,
			&rr.beneficiaryId,
			&rr.transferReference,
			&rr.createdBy,
			&rr.createdAt,
			&rr.modifiedBy,
			&rr.modifiedAt,
			&rr.version,
		); err != nil {
			log.Printf("Error processing records for account %d. Reason: %s", accountId, err)
			continue
		}

		if record, err = ledger.NewRecordFromRecord(rr); err != nil {
			log.Printf("Error loading record with id: %d from database. Reason: %s", rr.id, err)
			continue
		}

		entities = append(entities, record)
	}

	records := ledger.Records(entities)
	sort.Sort(records)

	return records, nil
}

func (d *DefaultRecordDao) GetRecordsForLastPeriod(ctx context.Context, accountId ledger.AccountId, tx *sql.Tx) (ledger.Records, error) {
	checkError := func(err error) bool {
		if err != nil {
			log.Printf("Error loading records for last period for account id: %d. Reason: %s", accountId, err)
			return true
		}
		return false
	}

	var rows *sql.Rows
	var err error
	if rows, err = d.db.QueryContext(ctx,
		`SELECT 
			MAX(r.date) 
		FROM 
			budget.record r 
		WHERE 
			r.account_id = $1`, accountId); checkError(err) {
		return ledger.Records{}, nil
	}

	defer rows.Close()
	rows.Next()

	var max sql.NullTime
	if err = rows.Scan(&max); max.Valid && checkError(err) {
		return ledger.Records{}, err
	}

	return d.GetRecordsForMonth(accountId, ledger.MakeCalendarMonthFromDate(max.Time))
}

func (d *DefaultRecordDao) GetRecordsForMonth(queryId ledger.AccountId, month ledger.CalendarMonth) (ledger.Records, error) {
	fromDate := month.FirstDay()
	toDate := month.LastDay()
	rows, err := d.db.Query(`
		SELECT 
			r.id, 
			r.category_id, 
			c.name,
			c.created_by,
			c.created_at,
			c.last_modified_by,
			c.last_modified_at,
			c.version,
			r.note, 
			r.currency, 
			r.amount_minor_units, 
			r.date, 
			r.type, 
			r.source_account_id,
			r.beneficiary_id,
			r.transfer_reference,
			r.created_by,
			r.created_at,
			r.last_modified_by,
			r.last_modified_at,
			r.version
		FROM 
			budget.record r 
		LEFT JOIN 
			budget.account a 
		ON 
			r.account_id = a.id 
		LEFT JOIN 
			budget.category c 
		ON 
			r.category_id = c.id 
		WHERE 
			a.id = $1 
			AND r.date >= $2 
			AND r.date <= $3 
		ORDER 
			BY r.date DESC`,
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

	entities := make([]ledger.Record, 0)
	for rows.Next() {

		var (
			rr     recordRecord
			record ledger.Record
		)

		if err = rows.Scan(
			&rr.id,
			&rr.category.id,
			&rr.category.name,
			&rr.category.createdBy,
			&rr.category.createdAt,
			&rr.category.modifiedBy,
			&rr.category.modifiedAt,
			&rr.category.version,
			&rr.note,
			&rr.currency,
			&rr.amountMinorUnits,
			&rr.date,
			&rr.recordType,
			&rr.sourceAccountId,
			&rr.beneficiaryId,
			&rr.transferReference,
			&rr.createdBy,
			&rr.createdAt,
			&rr.modifiedBy,
			&rr.modifiedAt,
			&rr.version,
		); err != nil {
			log.Printf("Error processing records for account %d. Reason: %s", queryId, err)
			continue
		}

		if record, err = ledger.NewRecordFromRecord(rr); err != nil {
			log.Printf("Error loading record with id: %d from database. Reason: %s", rr.id, err)
			continue
		}

		entities = append(entities, record)
	}

	records := ledger.Records(entities)
	sort.Sort(records)

	return records, nil
}
