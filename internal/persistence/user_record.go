package persistence

import (
	"database/sql"
	"log"
	"net/mail"
	"time"

	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
)

type userRecord struct {
	id         ledger.UserId
	email      string
	createdBy  string
	createdAt  time.Time
	modifiedBy sql.NullString
	modifiedAt sql.NullTime
	version    ledger.Version
}

func (ur userRecord) Id() ledger.UserId {
	return ur.id
}

func (ur userRecord) Email() *mail.Address {

	email, err := mail.ParseAddress(ur.email)
	if err != nil {
		log.Fatalf("Expected valid email to be saved in the database. Reason: %s", err)
	}
	return email
}

func (ur userRecord) CreatedBy() ledger.UpdatedBy {
	updatedBy, err := ledger.ParseUpdatedBy(ur.createdBy)
	if err != nil {
		log.Fatalf("Invalid createdBy persisted for record %d: %s", ur.id, ur.createdBy)
	}
	return updatedBy
}

func (ur userRecord) CreatedAtUTC() time.Time {
	return ur.createdAt
}

func (ur userRecord) ModifiedBy() ledger.UpdatedBy {
	if !ur.modifiedBy.Valid {
		return ledger.UpdatedBy{}
	}
	var (
		updatedBy ledger.UpdatedBy
		err       error
	)
	if updatedBy, err = ledger.ParseUpdatedBy(ur.modifiedBy.String); err != nil {
		log.Fatalf("Invalid modifiedBy persisted for record %d: %s", ur.id, ur.ModifiedBy())
	}
	return updatedBy
}

func (ur userRecord) ModifiedAtUTC() time.Time {
	if ur.modifiedAt.Valid {
		return ur.modifiedAt.Time
	}
	return time.Time{}
}

func (ur userRecord) Version() ledger.Version {
	return ur.version
}
