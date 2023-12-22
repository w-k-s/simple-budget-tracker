package ledger

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/w-k-s/simple-budget-tracker/pkg"
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
	var auditable Auditable = (*auditInfo)(nil)
	assert.Nil(suite.T(), auditable)
}

func (suite *AuditTestSuite) Test_GIVEN_editedByIsCreated_WHEN_userIdIsZero_THEN_errorIsReturned() {
	// GIVEN
	updatedBy, err := MakeUpdatedByUserId(0)

	// THEN
	assert.Equal(suite.T(), UpdatedBy{}, updatedBy)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), pkg.ErrAuditValidation, errorCode(err, 0))
	assert.Equal(suite.T(), "userId must be greater than 0", err.Error())
}

func (suite *AuditTestSuite) Test_GIVEN_auditableIsCreated_WHEN_createdAtIsNotProvided_THEN_errorIsReturned() {
	// GIVEN
	audit, err := makeAuditForModification(MustMakeUpdatedByUserId(1), time.Time{}, UpdatedBy{}, time.Time{}, 1)

	// THEN
	assert.Equal(suite.T(), auditInfo{}, audit)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), pkg.ErrAuditValidation, errorCode(err, 0))
	assert.Equal(suite.T(), "CreatedAt is required", err.Error())
}

func (suite *AuditTestSuite) Test_GIVEN_auditableIsCreated_WHEN_createdAtIsNotUTC_THEN_createdAtIsSetToUTCCorrectly() {
	// GIVEN
	createdAt := time.Date(2021, time.July, 28, 20, 10, 0, 0, time.FixedZone("UTC+4", 4*60*60))

	audit, err := makeAuditForModification(MustMakeUpdatedByUserId(1), createdAt, UpdatedBy{}, time.Time{}, 1)

	// THEN
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "UserId: 1", audit.CreatedBy().String())
	assert.Equal(suite.T(), "2021-07-28 16:10:00 +0000 UTC", audit.CreatedAtUTC().String())
	assert.Equal(suite.T(), UpdatedBy{}, audit.ModifiedBy())
	assert.Equal(suite.T(), time.Time{}, audit.ModifiedAtUTC())
	assert.Equal(suite.T(), Version(1), audit.Version())
}

func (suite *AuditTestSuite) Test_GIVEN_auditableIsModified_WHEN_modifiedAtIsNotInUTC_THEN_modifiedAtIsSetToUTCCorrectly() {
	// GIVEN
	createdAt := time.Date(2021, time.July, 28, 20, 10, 0, 0, time.FixedZone("UTC+4", 4*60*60))
	modifiedAt := createdAt.Add(time.Duration(1) * time.Hour)

	audit, err := makeAuditForModification(MustMakeUpdatedByUserId(1), createdAt, MustMakeUpdatedByUserId(2), modifiedAt, 2)

	// THEN
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "UserId: 1", audit.CreatedBy().String())
	assert.Equal(suite.T(), "2021-07-28 16:10:00 +0000 UTC", audit.CreatedAtUTC().String())
	assert.Equal(suite.T(), "UserId: 2", audit.ModifiedBy().String())
	assert.Equal(suite.T(), "2021-07-28 17:10:00 +0000 UTC", audit.ModifiedAtUTC().String())
	assert.Equal(suite.T(), Version(2), audit.Version())
}

func (suite *AuditTestSuite) Test_GIVEN_auditableIsCreated_WHEN_versionIsZero_THEN_errorIsReturned() {
	// GIVEN
	audit, err := makeAuditForModification(MustMakeUpdatedByUserId(1), time.Now(), UpdatedBy{}, time.Time{}, 0)

	// THEN
	assert.Equal(suite.T(), auditInfo{}, audit)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), pkg.ErrAuditValidation, errorCode(err, 0))
	assert.Equal(suite.T(), "Version must be greater than 0", err.Error())
}

func (suite *AuditTestSuite) Test_GIVEN_blankUpdatedByString_WHEN_parsed_THEN_errorIsReturned() {
	// GIVEN
	updatedBy, err := ParseUpdatedBy("")

	// THEN
	assert.Equal(suite.T(), UpdatedBy{}, updatedBy)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), pkg.ErrAuditUpdatedByBadFormat, errorCode(err, 0))
	assert.Equal(suite.T(), "Invalid createdBy/modifiedBy provided: \"\"", err.Error())
}

func (suite *AuditTestSuite) Test_GIVEN_blankPairsInUpdatedByString_WHEN_parsed_THEN_errorIsReturned() {
	// GIVEN
	updatedBy, err := ParseUpdatedBy(";")

	// THEN
	assert.Equal(suite.T(), UpdatedBy{}, updatedBy)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), pkg.ErrAuditUpdatedByBadFormat, errorCode(err, 0))
	assert.Equal(suite.T(), "Invalid createdBy/modifiedBy provided: \";\"", err.Error())
}

func (suite *AuditTestSuite) Test_GIVEN_pairWithOnlyKeyInUpdatedByString_WHEN_parsed_THEN_errorIsReturned() {
	// GIVEN
	updatedBy, err := ParseUpdatedBy(":")

	// THEN
	assert.Equal(suite.T(), UpdatedBy{}, updatedBy)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), pkg.ErrAuditUpdatedByBadFormat, errorCode(err, 0))
	assert.Equal(suite.T(), "Unknown createdBy/modifiedBy provided: \":\"", err.Error())
}

func (suite *AuditTestSuite) Test_GIVEN_userIdKeyWithInvalidUserIdInUpdatedByString_WHEN_parsed_THEN_errorIsReturned() {
	// GIVEN
	updatedBy, err := ParseUpdatedBy("UserId: Hello")

	// THEN
	assert.Equal(suite.T(), UpdatedBy{}, updatedBy)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), pkg.ErrAuditUpdatedByBadFormat, errorCode(err, 0))
	assert.Equal(suite.T(), "Invalid createdBy/modifiedBy provided: \"UserId: Hello\"", err.Error())
}

func (suite *AuditTestSuite) Test_GIVEN_userIdKeyWithBlankUserIdInUpdatedByString_WHEN_parsed_THEN_errorIsReturned() {
	// GIVEN
	updatedBy, err := ParseUpdatedBy("UserId: ")

	// THEN
	assert.Equal(suite.T(), UpdatedBy{}, updatedBy)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), pkg.ErrAuditUpdatedByBadFormat, errorCode(err, 0))
	assert.Equal(suite.T(), "Invalid createdBy/modifiedBy provided: \"UserId: \"", err.Error())
}

func (suite *AuditTestSuite) Test_GIVEN_userIdKeyWithValidUserIdInUpdatedByString_WHEN_parsed_THEN_updatedByIsReturned() {
	// GIVEN
	updatedBy, err := ParseUpdatedBy("UserId : 1")

	// THEN
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "UserId: 1", updatedBy.String())
}
