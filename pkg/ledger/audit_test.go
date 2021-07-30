package ledger

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AuditTestSuite struct {
	suite.Suite
}

func TestAuditTestSuite(t *testing.T) {
	suite.Run(t, new(AuditTestSuite))
}

// -- SUITE

func (suite *AuditTestSuite) Test_auditInfoIsAuditable() {
	// THEN
	var auditable Auditable = (*AuditInfo)(nil)
	assert.Nil(suite.T(), auditable)
}

func (suite *AuditTestSuite) Test_GIVEN_auditableIsCreated_WHEN_userIdIsZero_THEN_errorIsReturned() {
	// GIVEN
	audit, err := MakeAuditForCreation(0)

	// THEN
	assert.Equal(suite.T(), AuditInfo{}, audit)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrAuditValidation, err.(Error).Code())
	assert.Equal(suite.T(), "CreatedBy must be greater than 0", err.Error())
}

func (suite *AuditTestSuite) Test_GIVEN_auditableIsCreated_WHEN_createdAtIsNotProvided_THEN_errorIsReturned() {
	// GIVEN
	audit, err := MakeAuditForModification(UserId(1), time.Time{}, 0, time.Time{}, 1)

	// THEN
	assert.Equal(suite.T(), AuditInfo{}, audit)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrAuditValidation, err.(Error).Code())
	assert.Equal(suite.T(), "CreatedAt is required", err.Error())
}

func (suite *AuditTestSuite) Test_GIVEN_auditableIsCreated_WHEN_createdAtIsNotUTC_THEN_createdAtIsSetToUTCCorrectly() {
	// GIVEN
	createdAt := time.Date(2021, time.July, 28, 20, 10, 0, 0, time.FixedZone("UTC+4", 4*60*60))

	audit, err := MakeAuditForModification(UserId(1), createdAt, 0, time.Time{}, 1)

	// THEN
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), UserId(1), audit.CreatedBy())
	assert.Equal(suite.T(), "2021-07-28 16:10:00 +0000 UTC", audit.CreatedAtUTC().String())
	assert.Equal(suite.T(), UserId(0), audit.ModifiedBy())
	assert.Equal(suite.T(), time.Time{}, audit.ModifiedAtUTC())
	assert.Equal(suite.T(), Version(1), audit.Version())
}

func (suite *AuditTestSuite) Test_GIVEN_auditableIsModified_WHEN_modifiedAtIsNotInUTC_THEN_modifiedAtIsSetToUTCCorrectly() {
	// GIVEN
	createdAt := time.Date(2021, time.July, 28, 20, 10, 0, 0, time.FixedZone("UTC+4", 4*60*60))
	modifiedAt := createdAt.Add(time.Duration(1) * time.Hour)

	audit, err := MakeAuditForModification(UserId(1), createdAt, UserId(2), modifiedAt, 2)

	// THEN
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), UserId(1), audit.CreatedBy())
	assert.Equal(suite.T(), "2021-07-28 16:10:00 +0000 UTC", audit.CreatedAtUTC().String())
	assert.Equal(suite.T(), UserId(2), audit.ModifiedBy())
	assert.Equal(suite.T(), "2021-07-28 17:10:00 +0000 UTC", audit.ModifiedAtUTC().String())
	assert.Equal(suite.T(), Version(2), audit.Version())
}

func (suite *AuditTestSuite) Test_GIVEN_auditableIsCreated_WHEN_versionIsZero_THEN_errorIsReturned() {
	// GIVEN
	audit, err := MakeAuditForModification(UserId(1), time.Now(), 0, time.Time{}, 0)

	// THEN
	assert.Equal(suite.T(), AuditInfo{}, audit)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrAuditValidation, err.(Error).Code())
	assert.Equal(suite.T(), "Version must be greater than 0", err.Error())
}
