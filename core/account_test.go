package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AccountTestSuite struct {
	suite.Suite
}

func TestUserDaoTestSuite(t *testing.T) {
	suite.Run(t, new(AccountTestSuite))
}

// -- SUITE

func (suite *AccountTestSuite) Test_GIVEN_invalidAccountId_WHEN_AccoutnIsCreated_THEN_errorIsReturned() {
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

func (suite *AccountTestSuite) Test_GIVEN_emptyAccountName_WHEN_AccoutnIsCreated_THEN_errorIsReturned() {

	// WHEN
	account, err := NewAccount(2, "", "AED")

	// THEN
	assert.Nil(suite.T(), account)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrAccountValidation, err.(Error).Code())
	assert.Equal(suite.T(), "Name must be 1 and 255 characters long", err.(Error).Error())
	assert.Equal(suite.T(), "Name must be 1 and 255 characters long", err.(Error).Fields()["name"])
}

func (suite *AccountTestSuite) Test_GIVEN_noCurrency_WHEN_AccoutnIsCreated_THEN_errorIsReturned() {

	// WHEN
	account, err := NewAccount(2, "Main", "")

	// THEN
	assert.Nil(suite.T(), account)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrAccountValidation, err.(Error).Code())
	assert.Equal(suite.T(), "Currency must be 3 characters long", err.(Error).Error())
	assert.Equal(suite.T(), "Currency must be 3 characters long", err.(Error).Fields()["currency"])
}

func (suite *AccountTestSuite) Test_GIVEN_aNonExistantCurrency_WHEN_AccoutnIsCreated_THEN_errorIsReturned() {

	// WHEN
	account, err := NewAccount(2, "Main", "XXX")

	// THEN
	assert.Nil(suite.T(), account)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrAccountValidation, err.(Error).Code())
	assert.Equal(suite.T(), "No such currency \"XXX\"", err.(Error).Error())
	assert.Equal(suite.T(), "No such currency \"XXX\"", err.(Error).Fields()["currency"])
}

func (suite *AccountTestSuite) Test_GIVEN_validParameters_WHEN_AccoutnIsCreated_THEN_noErrorsAreReturned() {

	// WHEN
	account, err := NewAccount(2, "Main", "AED")

	// THEN
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), account)
	assert.Equal(suite.T(), account.Id(), AccountId(2))
	assert.Equal(suite.T(), account.Name(), "Main")
	assert.Equal(suite.T(), account.Currency(), "AED")
}
