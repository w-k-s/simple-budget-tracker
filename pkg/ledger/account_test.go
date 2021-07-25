package ledger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AccountTestSuite struct {
	suite.Suite
}

func TestAccountTestSuite(t *testing.T) {
	suite.Run(t, new(AccountTestSuite))
}

// -- SUITE

func (suite *AccountTestSuite) Test_GIVEN_invalidAccountId_WHEN_AccountIsCreated_THEN_errorIsReturned() {
	// GIVEN
	accountId := AccountId(0)

	// WHEN
	account, err := NewAccount(accountId, "test", "AED")

	// THEN
	assert.Nil(suite.T(), account)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrAccountValidation, err.(Error).Code())
	assert.Equal(suite.T(), "Id must be greater than 0", err.(Error).Error())
	assert.Equal(suite.T(), "Id must be greater than 0", err.(Error).Fields()["id"])
}

func (suite *AccountTestSuite) Test_GIVEN_emptyAccountName_WHEN_AccountIsCreated_THEN_errorIsReturned() {

	// WHEN
	account, err := NewAccount(2, "", "AED")

	// THEN
	assert.Nil(suite.T(), account)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrAccountValidation, err.(Error).Code())
	assert.Equal(suite.T(), "Name must be 1 and 25 characters long", err.(Error).Error())
	assert.Equal(suite.T(), "Name must be 1 and 25 characters long", err.(Error).Fields()["name"])
}

func (suite *AccountTestSuite) Test_GIVEN_noCurrency_WHEN_AccountIsCreated_THEN_errorIsReturned() {

	// WHEN
	account, err := NewAccount(2, "Main", "")

	// THEN
	assert.Nil(suite.T(), account)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrAccountValidation, err.(Error).Code())
	assert.Equal(suite.T(), "No such currency \"\", Currency must be 3 characters long", err.(Error).Error())
	assert.Equal(suite.T(), "No such currency \"\", Currency must be 3 characters long", err.(Error).Fields()["currency"])
}

func (suite *AccountTestSuite) Test_GIVEN_aNonExistantCurrency_WHEN_AccountIsCreated_THEN_errorIsReturned() {

	// WHEN
	account, err := NewAccount(2, "Main", "XXX")

	// THEN
	assert.Nil(suite.T(), account)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrAccountValidation, err.(Error).Code())
	assert.Equal(suite.T(), "No such currency \"XXX\"", err.(Error).Error())
	assert.Equal(suite.T(), "No such currency \"XXX\"", err.(Error).Fields()["currency"])
}

func (suite *AccountTestSuite) Test_GIVEN_validParameters_WHEN_AccountIsCreated_THEN_noErrorsAreReturned() {

	// WHEN
	account, err := NewAccount(2, "Main", "AED")

	// THEN
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), account)
	assert.Equal(suite.T(), AccountId(2), account.Id())
	assert.Equal(suite.T(), "Main", account.Name())
	assert.Equal(suite.T(), "AED", account.Currency())
}

func (suite *AccountTestSuite) Test_GIVEN_anAccount_WHEN_stringIsCalled_THEN_stringIsReadable() {

	// WHEN
	account, _ := NewAccount(2, "Main", "AED")

	// THEN
	assert.Equal(suite.T(), "Account{id: 2, name: Main, currency: AED}", account.String())
}