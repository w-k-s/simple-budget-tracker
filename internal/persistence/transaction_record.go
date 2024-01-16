package persistence

import (
	"database/sql"
	"log"
	"time"

	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
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
	beneficiaryType   sql.NullString
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

	category, err := ledger.NewCategoryFromRecord(rr.category)
	if err != nil {
		log.Fatalf("Failed to parse category from database for record id %d. Reason: %s", rr.id, err)
	}
	return category
}

func (rr recordRecord) Amount() ledger.Money {

	amount, err := ledger.NewMoney(rr.currency, rr.amountMinorUnits)
	if err != nil {
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

func (rr recordRecord) BeneficiaryType() ledger.AccountType {
	if rr.beneficiaryType.Valid {
		return ledger.AccountType(rr.beneficiaryType.String)
	}
	return ledger.NoBeneficiaryType
}

func (rr recordRecord) TransferReference() ledger.TransferReference {
	if rr.beneficiaryId.Valid {
		return ledger.TransferReference(rr.transferReference.String)
	}
	return ledger.NoTransferReference
}

func (rr recordRecord) CreatedBy() ledger.UpdatedBy {
	updatedBy, err := ledger.ParseUpdatedBy(rr.createdBy)
	if err != nil {
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
