package persistence

import (
	"database/sql"
	"log"
	"time"

	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
)

type categoryRecord struct {
	id         ledger.CategoryId
	name       string
	createdBy  string
	createdAt  time.Time
	modifiedBy sql.NullString
	modifiedAt sql.NullTime
	version    ledger.Version
}

func (cr categoryRecord) Id() ledger.CategoryId {
	return cr.id
}

func (cr categoryRecord) Name() string {
	return cr.name
}

func (cr categoryRecord) CreatedBy() ledger.UpdatedBy {
	updatedBy, err := ledger.ParseUpdatedBy(cr.createdBy)
	if err != nil {
		log.Fatalf("Invalid createdBy persisted for record %d: %s", cr.id, cr.createdBy)
	}
	return updatedBy
}

func (cr categoryRecord) CreatedAtUTC() time.Time {
	return cr.createdAt
}

func (cr categoryRecord) ModifiedBy() ledger.UpdatedBy {
	if !cr.modifiedBy.Valid {
		return ledger.UpdatedBy{}
	}
	var (
		updatedBy ledger.UpdatedBy
		err       error
	)
	if updatedBy, err = ledger.ParseUpdatedBy(cr.modifiedBy.String); err != nil {
		log.Fatalf("Invalid modifiedBy persisted for record %d: %s", cr.id, cr.ModifiedBy())
	}
	return updatedBy
}

func (cr categoryRecord) ModifiedAtUTC() time.Time {
	if cr.modifiedAt.Valid {
		return cr.modifiedAt.Time
	}
	return time.Time{}
}

func (cr categoryRecord) Version() ledger.Version {
	return cr.version
}
