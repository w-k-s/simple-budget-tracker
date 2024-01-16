package persistence

import (
	"database/sql"
	"log"
	"time"

	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
)

type accountRecord struct {
	id                       ledger.AccountId
	name                     string
	accountType              ledger.AccountType
	currency                 string
	currentBalanceMinorUnits sql.NullInt64
	createdBy                string
	createdAt                time.Time
	modifiedBy               sql.NullString
	modifiedAt               sql.NullTime
	version                  ledger.Version
}

func (ar accountRecord) Id() ledger.AccountId {
	return ar.id
}

func (ar accountRecord) Name() string {
	return ar.name
}

func (ar accountRecord) Type() ledger.AccountType {
	return ar.accountType
}

func (ar accountRecord) Currency() string {
	return ar.currency
}

func (ar accountRecord) CurrentBalanceMinorUnits() int64 {
	if ar.currentBalanceMinorUnits.Valid {
		return ar.currentBalanceMinorUnits.Int64
	}
	return 0
}

func (ar accountRecord) CreatedBy() ledger.UpdatedBy {
	updatedBy, err := ledger.ParseUpdatedBy(ar.createdBy)
	if err != nil {
		log.Fatalf("Invalid createdBy persisted for record %d: %s", ar.id, ar.createdBy)
	}
	return updatedBy
}

func (ar accountRecord) CreatedAtUTC() time.Time {
	return ar.createdAt
}

func (ar accountRecord) ModifiedBy() ledger.UpdatedBy {
	if !ar.modifiedBy.Valid {
		return ledger.UpdatedBy{}
	}
	var (
		updatedBy ledger.UpdatedBy
		err       error
	)
	if updatedBy, err = ledger.ParseUpdatedBy(ar.modifiedBy.String); err != nil {
		log.Fatalf("Invalid modifiedBy persisted for record %d: %s", ar.id, ar.ModifiedBy())
	}
	return updatedBy
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
