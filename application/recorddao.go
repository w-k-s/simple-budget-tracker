package application

import (
	"database/sql"
	"fmt"
	"log"

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
	// TODO
	return fmt.Errorf("To implement")
}

func (d *DefaultRecordDao) SaveTx(accountId core.AccountId, r *core.Record, tx *sql.Tx) error {
	// TODO
	return fmt.Errorf("To implement")
}

func (d *DefaultRecordDao) Search(id core.AccountId, search core.RecordSearch) (core.Records, error) {
	// TODO
	return nil, fmt.Errorf("To implement")
}

func (d *DefaultRecordDao) GetRecordsForLastPeriod(accountId core.AccountId) (core.Records, error) {
	// TODO
	return nil, fmt.Errorf("To implement")
}

func (d *DefaultRecordDao) GetRecordsForMonth(id core.AccountId, month int, year int) (core.Records, error) {
	// TODO
	return nil, fmt.Errorf("To implement")
}
