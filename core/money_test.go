package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MoneyTestSuite struct {
	suite.Suite
}

func TestMoneyDaoTestSuite(t *testing.T) {
	suite.Run(t, new(MoneyTestSuite))
}

// -- SUITE

func (suite *AccountTestSuite) Test_GIVEN_blankCurrencyCode_WHEN_AmountIsCreated_THEN_errorIsReturned() {
	// WHEN
	money, err := NewMoney("", 1000)

	// THEN
	assert.Nil(suite.T(), money)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), uint64(1010), uint64(err.(Error).Code()))
	assert.Equal(suite.T(), "invalid currency code \"\"", err.(Error).Error())
	assert.Equal(suite.T(), "", err.(Error).Fields()["code"])
}

func (suite *AccountTestSuite) Test_GIVEN_invalidCurrency_WHEN_currencyIsCreated_THEN_errorIsReturned() {

	// WHEN
	money, err := NewMoney("III", 1000)

	// THEN
	assert.Nil(suite.T(), money)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), uint64(1010), uint64(err.(Error).Code()))
	assert.Equal(suite.T(), "invalid currency code \"III\"", err.(Error).Error())
	assert.Equal(suite.T(), "III", err.(Error).Fields()["code"])
}

func (suite *AccountTestSuite) Test_GIVEN_negativeAmount_WHEN_moneyIsCreated_THEN_amountIsNegative() {

	// WHEN
	money, err := NewMoney("AED", -1000)

	// THEN
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), "AED -10.00", money.String())
}

func (suite *AccountTestSuite) Test_GIVEN_positiveAmount_WHEN_moneyIsCreated_THEN_amountIsPositive() {

	// WHEN
	money, err := NewMoney("AED", 1000)

	// THEN
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), "AED 10.00", money.String())
}

func (suite *AccountTestSuite) Test_GIVEN_zeroAmount_WHEN_moneyIsCreated_THEN_amountIsPositive() {

	// WHEN
	money, err := NewMoney("AED", 0)

	// THEN
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), "AED 0.00", money.String())
}
