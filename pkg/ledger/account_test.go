package ledger

import (
	"sort"
	"testing"
	"time"

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
	account, err := NewAccount(accountId, "test", Current, "AED", MustMakeUpdatedByUserId(UserId(1)))

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), Account{}, account)
	assert.Equal(suite.T(), ErrAccountValidation, err.(Error).Code())
	assert.Equal(suite.T(), "Id must be greater than 0", err.(Error).Error())
	assert.Equal(suite.T(), "Id must be greater than 0", err.(Error).Fields()["id"])
}

func (suite *AccountTestSuite) Test_GIVEN_emptyAccountName_WHEN_AccountIsCreated_THEN_errorIsReturned() {

	// WHEN
	account, err := NewAccount(2, "", Current, "AED", MustMakeUpdatedByUserId(UserId(1)))

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), Account{}, account)
	assert.Equal(suite.T(), ErrAccountValidation, err.(Error).Code())
	assert.Equal(suite.T(), "Name must be 1 and 25 characters long", err.(Error).Error())
	assert.Equal(suite.T(), "Name must be 1 and 25 characters long", err.(Error).Fields()["name"])
}

func (suite *AccountTestSuite) Test_GIVEN_noCurrency_WHEN_AccountIsCreated_THEN_errorIsReturned() {

	// WHEN
	account, err := NewAccount(2, "Main", Current, "", MustMakeUpdatedByUserId(UserId(1)))

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), Account{}, account)
	assert.Equal(suite.T(), ErrAccountValidation, err.(Error).Code())
	assert.Equal(suite.T(), "currency is required", err.(Error).Error())
	assert.Equal(suite.T(), "currency is required", err.(Error).Fields()["currency"])
}

func (suite *AccountTestSuite) Test_GIVEN_aNonExistantCurrency_WHEN_AccountIsCreated_THEN_errorIsReturned() {

	// WHEN
	account, err := NewAccount(2, "Main", Current, "XXX", MustMakeUpdatedByUserId(UserId(1)))

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), Account{}, account)
	assert.Equal(suite.T(), ErrAccountValidation, err.(Error).Code())
	assert.Equal(suite.T(), "No such currency 'XXX'", err.(Error).Error())
	assert.Equal(suite.T(), "No such currency 'XXX'", err.(Error).Fields()["currency"])
}

func (suite *AccountTestSuite) Test_GIVEN_validParameters_WHEN_AccountIsCreated_THEN_noErrorsAreReturned() {

	// WHEN
	account, err := NewAccount(2, "Main", Current, "AED", MustMakeUpdatedByUserId(UserId(1)))

	// THEN
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), account)
	assert.Equal(suite.T(), AccountId(2), account.Id())
	assert.Equal(suite.T(), "Main", account.Name())
	assert.Equal(suite.T(), "AED", account.Currency())
	assert.Equal(suite.T(), "AED 0.00", account.CurrentBalance().String())
	assert.Equal(suite.T(), "UserId: 1", account.CreatedBy().String())
	assert.Equal(suite.T(), Version(1), account.Version())
	assert.True(suite.T(), time.Now().UTC().Sub(account.CreatedAtUTC()) < time.Duration(1)*time.Second)
}

func (suite *AccountTestSuite) Test_GIVEN_anAccount_WHEN_stringIsCalled_THEN_stringIsReadable() {

	// WHEN
	account, _ := NewAccount(2, "Main", Current, "AED", MustMakeUpdatedByUserId(UserId(1)))

	// THEN
	assert.Equal(suite.T(), "Account{id: 2, name: Main, type: Current, currency: AED, balance: AED 0.00}", account.String())
}

func (suite *AccountTestSuite) Test_GIVEN_accounts_WHEN_namesIsCalled_THEN_sliceOfSortedAccountNamesIsReturned() {

	// WHEN
	account1, _ := NewAccount(1, "Current", Current, "AED", MustMakeUpdatedByUserId(UserId(1)))
	account2, _ := NewAccount(1, "Savings", Saving, "AED", MustMakeUpdatedByUserId(UserId(1)))

	accounts := Accounts{
		account1,
		account2,
	}

	// THEN
	assert.Equal(suite.T(), []string{"Current", "Savings"}, accounts.Names())
}

func (suite *AccountTestSuite) Test_GIVEN_accounts_WHEN_sortIsCalled_THEN_accountsAreSortedInPlace() {

	// WHEN
	account1, _ := NewAccount(1, "Current", Current, "AED", MustMakeUpdatedByUserId(UserId(1)))
	account2, _ := NewAccount(1, "Savings",Saving, "AED", MustMakeUpdatedByUserId(UserId(1)))

	accounts := Accounts{
		account2,
		account1,
	}
	sort.Sort(accounts)

	// THEN
	assert.Equal(suite.T(), "Current", accounts[0].Name())
	assert.Equal(suite.T(), "Savings", accounts[1].Name())
}

func (suite *AccountTestSuite) Test_GIVEN_accounts_WHEN_stringIsCalled_THEN_stringOfEachAccountIsPrintedInAlphabeticalOrder() {

	// WHEN
	account1, _ := NewAccount(1, "Current", Current, "AED", MustMakeUpdatedByUserId(UserId(1)))
	account2, _ := NewAccount(2, "Savings", Saving, "AED", MustMakeUpdatedByUserId(UserId(1)))

	accounts := Accounts{
		account2,
		account1,
	}

	// THEN
	assert.Equal(suite.T(), "Accounts{Account{id: 1, name: Current, type: Current, currency: AED, balance: AED 0.00}, Account{id: 2, name: Savings, type: Saving, currency: AED, balance: AED 0.00}}", accounts.String())
}

func (suite *AccountTestSuite) Test_GIVEN_anAccountName_WHEN_accountIsCreated_THEN_accountNameIsCapitalized() {

	// WHEN
	account1, _ := NewAccount(1, "current", Current, "AED", MustMakeUpdatedByUserId(UserId(1)))
	account2, _ := NewAccount(2, "sAVINGS", Saving, "AED", MustMakeUpdatedByUserId(UserId(1)))

	accounts := Accounts{
		account2,
		account1,
	}

	// THEN
	assert.Equal(suite.T(), "Accounts{Account{id: 1, name: Current, type: Current, currency: AED, balance: AED 0.00}, Account{id: 2, name: Savings, type: Saving, currency: AED, balance: AED 0.00}}", accounts.String())
}
