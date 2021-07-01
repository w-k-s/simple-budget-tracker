package application

import (
	"log"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/w-k-s/simple-budget-tracker/core"
)

type AccountDaoTestSuite struct {
	suite.Suite
	userDao    core.UserDao
	accountDao core.AccountDao
}

func TestAccountDaoTestSuite(t *testing.T) {
	suite.Run(t, new(AccountDaoTestSuite))
}

// -- SETUP

const (
	testUserId    = core.UserId(1)
	testUserEmail = "jack.torrence@theoverlook.com"
)

func (suite *AccountDaoTestSuite) SetupTest() {
	suite.userDao = UserDao
	suite.accountDao = AccountDao

	aUser, _ := core.NewUserWithEmailString(testUserId, testUserEmail)
	if err := suite.userDao.Save(aUser); err != nil {
		log.Fatalf("AccountDaoTestSuite: Test setup failed: %s", err)
	}
}

func (suite *AccountDaoTestSuite) TearDownTest() {
	if err := ClearTables(); err != nil {
		log.Fatalf("Failed to tear down AccountDaoTestSuite: %s", err)
	}
}

// -- SUITE

func (suite *AccountDaoTestSuite) Test_WHEN_NewAccountIdIsCalled_THEN_accountIdIsReturnedFromDatabaseSequence() {
	// WHEN
	accountId, err := suite.accountDao.NewAccountId()

	// THEN
	assert.Nil(suite.T(), err)
	assert.Positive(suite.T(), accountId)
}

func (suite *AccountDaoTestSuite) Test_Given_anAccount_WHEN_theAccountIsSaved_THEN_accountCanBeRetrievedById() {
	// GIVEN
	anAccount, _ := core.NewAccount(core.AccountId(1), "Current", "AED")

	// WHEN
	_ = suite.accountDao.Save(testUserId, anAccount)
	theAccount, err := suite.accountDao.GetAccountById(core.AccountId(1))

	// THEN
	assert.Nil(suite.T(), err)
	assert.EqualValues(suite.T(), core.AccountId(1), theAccount.Id())
	assert.EqualValues(suite.T(), "Current", theAccount.Name())
	assert.EqualValues(suite.T(), "AED", theAccount.Currency())
}

func (suite *AccountDaoTestSuite) Test_Given_anAccount_WHEN_theAccountIsSaved_THEN_accountCanBeRetrievedByUserId() {
	// GIVEN
	currentAccount, _ := core.NewAccount(1, "Current", "AED")
	lifeSavingsAccount, _ := core.NewAccount(2, "Life Savings", "EUR")

	// WHEN
	_ = suite.accountDao.Save(testUserId, currentAccount)
	_ = suite.accountDao.Save(testUserId, lifeSavingsAccount)
	allAccounts, err := suite.accountDao.GetAccountsByUserId(testUserId)

	// THEN
	assert.Nil(suite.T(), err)
	assert.Len(suite.T(), allAccounts, 2)

	assert.EqualValues(suite.T(), core.AccountId(1), allAccounts[0].Id())
	assert.EqualValues(suite.T(), "Current", allAccounts[0].Name())
	assert.EqualValues(suite.T(), "AED", allAccounts[0].Currency())

	assert.EqualValues(suite.T(), core.AccountId(2), allAccounts[1].Id())
	assert.EqualValues(suite.T(), "Life Savings", allAccounts[1].Name())
	assert.EqualValues(suite.T(), "EUR", allAccounts[1].Currency())
}

func (suite *AccountDaoTestSuite) Test_Given_anAccountId_WHEN_noAccountWithThatIdExists_THEN_appropriateErrorIsReturned() {
	// GIVEN
	accountId := core.AccountId(1)

	// WHEN
	theAccount, err := suite.accountDao.GetAccountById(accountId)

	// THEN
	assert.Nil(suite.T(), theAccount)

	coreError := err.(core.Error)
	assert.EqualValues(suite.T(), core.ErrAccountNotFound, uint64(coreError.Code()))
	assert.EqualValues(suite.T(), "Account with id 1 not found", coreError.Error())
}

func (suite *AccountDaoTestSuite) Test_Given_twoAccounts_WHEN_theAccountsHaveTheSameName_THEN_onlyOneAccountIsSaved() {
	// GIVEN
	account1, _ := core.NewAccount(1, "Current", "AED")
	account2, _ := core.NewAccount(2, "Current", "AED")

	// WHEN
	err1 := suite.accountDao.Save(testUserId, account1)
	err2 := suite.accountDao.Save(testUserId, account2)

	// THEN
	assert.Nil(suite.T(), err1)
	assert.NotNil(suite.T(), err2)

	coreError := err2.(core.Error)
	assert.Equal(suite.T(), core.ErrAccountNameDuplicated, coreError.Code())
	assert.Equal(suite.T(), "Account name must be unique", coreError.Error())
}
