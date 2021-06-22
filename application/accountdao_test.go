package application

import (
	"context"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/w-k-s/simple-budget-tracker/core"
	"github.com/w-k-s/simple-budget-tracker/migrations"
)

type AccountDaoTestSuite struct {
	suite.Suite
	containerCtx context.Context
	postgresC    tc.Container
	userDao      core.UserDao
	accountDao   core.AccountDao
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
	containerCtx, postgresC, dataSourceName, err := requestPostgresTestContainer()
	if err != nil {
		panic(err)
	}

	suite.containerCtx = *containerCtx
	suite.postgresC = postgresC
	migrations.MustRunMigrations(TestContainerDriverName, dataSourceName, os.Getenv("TEST_MIGRATIONS_DIRECTORY"))

	suite.userDao = MustOpenUserDao(TestContainerDriverName, dataSourceName)
	suite.accountDao = MustOpenAccountDao(TestContainerDriverName, dataSourceName)

	aUser, _ := core.NewUserWithEmailString(testUserId, testUserEmail)
	if err = suite.userDao.Save(aUser); err != nil {
		log.Fatalf("AccountDaoTestSuite: Test setup failed: %s", err)
	}
}

// -- TEARDOWN

func (suite *AccountDaoTestSuite) TearDownTest() {
	if container := suite.postgresC; container != nil {
		_ = container.Terminate(suite.containerCtx)
	}
	suite.accountDao.Close()
	suite.userDao.Close()
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
	anAccount, _ := core.NewAccount(1, "Current", "AED")

	// WHEN
	_ = suite.accountDao.Save(testUserId, anAccount)
	theAccount, err := suite.accountDao.GetAccountById(1)

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
	assert.EqualValues(suite.T(), uint64(1008), uint64(coreError.Code()))
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
	assert.Equal(suite.T(), uint64(1009), uint64(coreError.Code()))
	assert.Equal(suite.T(), "Account name must be unique", coreError.Error())
}
