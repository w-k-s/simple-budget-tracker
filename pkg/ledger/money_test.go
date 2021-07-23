package ledger

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

func (suite *MoneyTestSuite) Test_GIVEN_negativeAmount_WHEN_amountSignChecksAreDone_THEN_resultsAreCorrect() {
	money, _ := NewMoney("AED", -1)
	assert.True(suite.T(), money.IsNegative())
	assert.False(suite.T(), money.IsPositive())
	assert.False(suite.T(), money.IsZero())
}

func (suite *MoneyTestSuite) Test_GIVEN_positiveAmount_WHEN_amountSignChecksAreDone_THEN_resultsAreCorrect() {
	money, _ := NewMoney("AED", 1)
	assert.False(suite.T(), money.IsNegative())
	assert.True(suite.T(), money.IsPositive())
	assert.False(suite.T(), money.IsZero())
}

func (suite *MoneyTestSuite) Test_GIVEN_zeroAmount_WHEN_amountSignChecksAreDone_THEN_resultsAreCorrect() {
	money, _ := NewMoney("AED", 0)
	assert.False(suite.T(), money.IsNegative())
	assert.False(suite.T(), money.IsPositive())
	assert.True(suite.T(), money.IsZero())
}

func (suite *MoneyTestSuite) Test_GIVEN_amount_WHEN_itIsAbsoluted_THEN_amountIsPositive() {

	// GIVEN
	negativeAmount, _ := NewMoney("AED", -1)
	positiveAmount, _ := NewMoney("AED", -1)
	zeroAmount, _ := NewMoney("AED", 0)

	// WHEN
	absOfNegative, _ := negativeAmount.Abs()
	absOfPositive, _ := positiveAmount.Abs()
	absOfZero, _ := zeroAmount.Abs()

	assert.Equal(suite.T(), "AED 0.01", absOfNegative.String())
	assert.Equal(suite.T(), "AED 0.01", absOfPositive.String())
	assert.Equal(suite.T(), "AED 0.00", absOfZero.String())
}

func (suite *MoneyTestSuite) Test_GIVEN_amount_WHEN_itIsNegated_THEN_amountIsNegative() {

	// GIVEN
	negativeAmount, _ := NewMoney("AED", -1)
	positiveAmount, _ := NewMoney("AED", -1)
	zeroAmount, _ := NewMoney("AED", 0)

	// WHEN
	negativeNegated, _ := negativeAmount.Negate()
	positiveNegated, _ := positiveAmount.Negate()
	zeroNegated, _ := zeroAmount.Negate()

	assert.Equal(suite.T(), "AED -0.01", negativeNegated.String())
	assert.Equal(suite.T(), "AED -0.01", positiveNegated.String())
	assert.Equal(suite.T(), "AED 0.00", zeroNegated.String())
}

func (suite *MoneyTestSuite) Test_GIVEN_aMoney_WHEN_stringIsCalled_THEN_stringIsReadable() {

	// WHEN
	money, _ := NewMoney("AED", 10000)

	// THEN
	assert.Equal(suite.T(), "AED 100.00", money.String())
}

func (suite *MoneyTestSuite) Test_GIVEN_amountsOfSameCurrency_WHEN_adding_THEN_sumIsCalculatedCorrectly() {

	// GIVEN
	money1, _ := NewMoney("AED", 2975)
	money2, _ := NewMoney("AED", 21644)

	// WHEN
	total, err := money1.Add(money2)

	// THEN
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "AED 246.19", total.String())
}

func (suite *MoneyTestSuite) Test_GIVEN_amountsOfDifferentCurrency_WHEN_adding_THEN_errorIsRetuend() {

	// GIVEN
	money1, _ := NewMoney("AED", 2975)
	money2, _ := NewMoney("KWD", 21644)

	// WHEN
	total, err := money1.Add(money2)

	// THEN
	assert.Nil(suite.T(), total)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrAmountMismatchingCurrencies, err.(Error).Code())
	assert.Equal(suite.T(), "Can not sum mismatching currencies", err.(Error).Error())
}
