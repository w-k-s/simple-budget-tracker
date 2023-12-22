package ledger

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/w-k-s/simple-budget-tracker/pkg"
)

type UpdatedBy struct {
	value string
}

func ParseUpdatedBy(updatedBy string) (UpdatedBy, error) {
	parts := strings.Split(updatedBy, ";")
	for _, pairs := range parts {
		pair := strings.Split(pairs, ":")
		if len(pair) != 2 {
			return UpdatedBy{}, pkg.ValidationErrorWithFields(pkg.ErrAuditUpdatedByBadFormat, fmt.Sprintf("Invalid createdBy/modifiedBy provided: %q", updatedBy), nil, nil)
		}
		key := strings.TrimSpace(pair[0])
		value := strings.TrimSpace(pair[1])
		if key == "UserId" {
			var (
				userId int
				err    error
			)
			if userId, err = strconv.Atoi(value); err != nil {
				return UpdatedBy{}, pkg.ValidationErrorWithFields(pkg.ErrAuditUpdatedByBadFormat, fmt.Sprintf("Invalid createdBy/modifiedBy provided: %q", updatedBy), err, nil)
			}
			return MakeUpdatedByUserId(UserId(userId))
		}
	}
	return UpdatedBy{}, pkg.ValidationErrorWithFields(pkg.ErrAuditUpdatedByBadFormat, fmt.Sprintf("Unknown createdBy/modifiedBy provided: %q", updatedBy), nil, nil)
}

func (u UpdatedBy) String() string {
	return u.value
}

func MakeUpdatedByUserId(userId UserId) (UpdatedBy, error) {
	if userId <= 0 {
		return UpdatedBy{}, pkg.ValidationErrorWithFields(pkg.ErrAuditValidation, "userId must be greater than 0", nil, nil)
	}
	return UpdatedBy{fmt.Sprintf("UserId: %d", userId)}, nil
}

func MustMakeUpdatedByUserId(userId UserId) UpdatedBy {
	var (
		updatedBy UpdatedBy
		err       error
	)
	if updatedBy, err = MakeUpdatedByUserId(userId); err != nil {
		log.Fatalf("Invalid userId provided for createdBy/modifiedBy. Reason: %s", err)
	}
	return updatedBy
}

// TODO: MakeEditedByTask -> "CronTask: taskName"
// TODO: MakeEditedByImport -> "Import: fileName"

type Version uint64

type Auditable interface {
	CreatedBy() UpdatedBy
	CreatedAtUTC() time.Time
	ModifiedBy() UpdatedBy
	ModifiedAtUTC() time.Time
	Version() Version
}

type auditInfo struct {
	createdBy     UpdatedBy
	createdAtUTC  time.Time
	modifiedBy    UpdatedBy
	modifiedAtUTC time.Time
	version       Version
}

func (ai auditInfo) CreatedBy() UpdatedBy {
	return ai.createdBy
}

func (ai auditInfo) CreatedAtUTC() time.Time {
	return ai.createdAtUTC
}

func (ai auditInfo) ModifiedBy() UpdatedBy {
	return ai.modifiedBy
}

func (ai auditInfo) ModifiedAtUTC() time.Time {
	return ai.modifiedAtUTC
}

func (ai auditInfo) Version() Version {
	return ai.version
}

func makeAuditForCreation(updatedBy UpdatedBy) (auditInfo, error) {
	return makeAuditForModification(
		updatedBy,
		time.Now().UTC(),
		UpdatedBy{},
		time.Time{},
		1,
	)
}

func makeAuditForModification(
	createdBy UpdatedBy,
	createdAt time.Time,
	modifiedBy UpdatedBy,
	modifiedAt time.Time,
	version Version,
) (auditInfo, error) {

	errors := validate.Validate(
		&validators.IntIsGreaterThan{Name: "Version", Field: int(version), Compared: 0, Message: "Version must be greater than 0"},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: createdAt, Message: "CreatedAt is required"},
	)

	var err error
	if err = pkg.ValidationErrorWithErrors(pkg.ErrAuditValidation, "", errors); err != nil {
		return auditInfo{}, err
	}

	var modifiedAtUTC time.Time
	var epoch time.Time

	if epoch.UnixNano() != modifiedAt.UnixNano() {
		modifiedAtUTC = modifiedAt.In(time.UTC)
	}

	return auditInfo{
		createdBy:     createdBy,
		createdAtUTC:  createdAt.In(time.UTC),
		modifiedBy:    modifiedBy,
		modifiedAtUTC: modifiedAtUTC,
		version:       version,
	}, nil
}
