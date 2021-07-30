package ledger

import (
	"time"

	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type Version uint64

type Auditable interface {
	CreatedBy() UserId
	CreatedAtUTC() time.Time
	ModifiedBy() UserId
	ModifiedAtUTC() time.Time
	Version() Version
}

type AuditInfo struct {
	createdBy     UserId
	createdAtUTC  time.Time
	modifiedBy    UserId
	modifiedAtUTC time.Time
	version       Version
}

func (ai AuditInfo) CreatedBy() UserId {
	return ai.createdBy
}

func (ai AuditInfo) CreatedAtUTC() time.Time {
	return ai.createdAtUTC
}

func (ai AuditInfo) ModifiedBy() UserId {
	return ai.modifiedBy
}

func (ai AuditInfo) ModifiedAtUTC() time.Time {
	return ai.modifiedAtUTC
}

func (ai AuditInfo) Version() Version {
	return ai.version
}

func MakeAuditForCreation(userId UserId) (AuditInfo, error) {
	return MakeAuditForModification(
		userId,
		time.Now().UTC(),
		0,
		time.Time{},
		1,
	)
}

func MakeAuditForModification(
	createdBy UserId,
	createdAt time.Time,
	modifiedBy UserId,
	modifiedAt time.Time,
	version Version,
) (AuditInfo, error) {

	errors := validate.Validate(
		&validators.IntIsGreaterThan{Name: "CreatedBy", Field: int(createdBy), Compared: 0, Message: "CreatedBy must be greater than 0"},
		&validators.IntIsGreaterThan{Name: "Version", Field: int(version), Compared: 0, Message: "Version must be greater than 0"},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: createdAt, Message: "CreatedAt is required"},
	)

	var err error
	if err = makeCoreValidationError(ErrAuditValidation, errors); err != nil {
		return AuditInfo{}, err
	}

	var modifiedAtUTC time.Time
	var epoch time.Time

	if epoch.UnixNano() != modifiedAt.UnixNano() {
		modifiedAtUTC = modifiedAt.In(time.UTC)
	}

	return AuditInfo{
		createdBy:     createdBy,
		createdAtUTC:  createdAt.In(time.UTC),
		modifiedBy:    modifiedBy,
		modifiedAtUTC: modifiedAtUTC,
		version:       version,
	}, nil
}
