package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MoneyTestSuite struct {
	suite.Suite
}

func TestMoneyTestSuite(t *testing.T) {
	suite.Run(t, new(MoneyTestSuite))
}

// -- SUITE

func (suite *MoneyTestSuite) Test_GIVEN_blankCurrencyCode_WHEN_AmountIsCreated_THEN_errorIsReturned() {
	// WHEN
	money, err := NewMoney("", 1000)

	// THEN
	assert.Nil(suite.T(), money)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrCurrencyInvalidCode, err.(Error).Code())
	assert.Equal(suite.T(), "invalid currency code \"\"", err.(Error).Error())
	assert.Equal(suite.T(), "", err.(Error).Fields()["code"])
}

func (suite *MoneyTestSuite) Test_GIVEN_invalidCurrency_WHEN_currencyIsCreated_THEN_errorIsReturned() {

	// WHEN
	money, err := NewMoney("III", 1000)

	// THEN
	assert.Nil(suite.T(), money)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrCurrencyInvalidCode, err.(Error).Code())
	assert.Equal(suite.T(), "invalid currency code \"III\"", err.(Error).Error())
	assert.Equal(suite.T(), "III", err.(Error).Fields()["code"])
}

func (suite *MoneyTestSuite) Test_GIVEN_negativeAmount_WHEN_moneyIsCreated_THEN_amountIsNegative() {

	// WHEN
	money, err := NewMoney("AED", -1000)

	// THEN
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), "AED -10.00", money.String())
}

func (suite *MoneyTestSuite) Test_GIVEN_positiveAmount_WHEN_moneyIsCreated_THEN_amountIsPositive() {

	// WHEN
	money, err := NewMoney("AED", 1000)

	// THEN
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), "AED 10.00", money.String())
}

func (suite *MoneyTestSuite) Test_GIVEN_zeroAmount_WHEN_moneyIsCreated_THEN_amountIsPositive() {

	// WHEN
	money, err := NewMoney("AED", 0)

	// THEN
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), "AED 0.00", money.String())
}

func (suite *MoneyTestSuite) Test_GIVEN_currenciesWithDifferentDenominations_WHEN_moneyIsCreated_THEN_minorUnitsConvertedToDecimalCorrectly() {
	quickMoney := func(currency string, amountMinorUnits int64) Money {
		money, _ := NewMoney(currency, amountMinorUnits)
		return money
	}

	assert.Equal(suite.T(), "AED 1.00", quickMoney("AED", 100).String())
	assert.Equal(suite.T(), "JPY 100", quickMoney("JPY", 100).String())
	assert.Equal(suite.T(), "KWD 0.100", quickMoney("KWD", 100).String())
}

func (suite *MoneyTestSuite) Test_GIVEN_aMoney_WHEN_stringIsCalled_THEN_stringIsReadable() {

	// WHEN
	money, _ := NewMoney("AED", 10000)

	// THEN
	assert.Equal(suite.T(), "AED 100.00", money.String())
}
