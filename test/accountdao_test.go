package test

import (
	"context"
	"log"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
	dao "github.com/w-k-s/simple-budget-tracker/pkg/persistence"
)

type AccountDaoTestSuite struct {
	suite.Suite
	userDao    dao.UserDao
	accountDao dao.AccountDao
	testUser   ledger.User
}

func TestAccountDaoTestSuite(t *testing.T) {
	suite.Run(t, new(AccountDaoTestSuite))
}

// -- SETUP

func (suite *AccountDaoTestSuite) SetupTest() {
	aUser, _ := ledger.NewUserWithEmailString(1, "jack.torrence@theoverlook.com")
	if err := UserDao.Save(aUser); err != nil {
		log.Fatalf("AccountDaoTestSuite: Test setup failed: %s", err)
	}

	suite.userDao = UserDao
	suite.accountDao = AccountDao
	suite.testUser = aUser
}

func (suite *AccountDaoTestSuite) TearDownTest() {
	if err := ClearTables(); err != nil {
		log.Fatalf("Failed to tear down AccountDaoTestSuite: %s", err)
	}
}

// -- SUITE

func (suite *AccountDaoTestSuite) Test_WHEN_NewAccountIdIsCalled_THEN_accountIdIsReturnedFromDatabaseSequence() {
	// WHEN
	tx := suite.accountDao.MustBeginTx()
	accountId, err := suite.accountDao.NewAccountId(tx)
	_ = tx.Commit()

	// THEN
	assert.Nil(suite.T(), err)
	assert.Positive(suite.T(), accountId)
}

func (suite *AccountDaoTestSuite) Test_Given_anAccount_WHEN_theAccountIsSaved_THEN_accountCanBeRetrievedByUserId() {
	// GIVEN
	currentAccount, _ := ledger.NewAccount(1, "Current", "AED", ledger.MustMakeUpdatedByUserId(suite.testUser.Id()))
	lifeSavingsAccount, _ := ledger.NewAccount(2, "Life Savings", "EUR", ledger.MustMakeUpdatedByUserId(suite.testUser.Id()))

	// WHEN
	tx := suite.accountDao.MustBeginTx()
	_ = suite.accountDao.SaveTx(context.Background(), suite.testUser.Id(), ledger.Accounts{currentAccount, lifeSavingsAccount}, tx)
	_ = tx.Commit()

	tx = suite.accountDao.MustBeginTx()
	allAccounts, err := suite.accountDao.GetAccountsByUserId(context.Background(), suite.testUser.Id(), tx)
	_ = tx.Commit()

	// THEN
	assert.Nil(suite.T(), err)
	assert.Len(suite.T(), allAccounts, 2)

	assert.EqualValues(suite.T(), ledger.AccountId(1), allAccounts[0].Id())
	assert.EqualValues(suite.T(), "Current", allAccounts[0].Name())
	assert.EqualValues(suite.T(), "AED", allAccounts[0].Currency())

	assert.EqualValues(suite.T(), ledger.AccountId(2), allAccounts[1].Id())
	assert.EqualValues(suite.T(), "Life Savings", allAccounts[1].Name())
	assert.EqualValues(suite.T(), "EUR", allAccounts[1].Currency())
}

func (suite *AccountDaoTestSuite) Test_Given_anAccount_WHEN_theAccountIsSaved_THEN_accountCanBeRetrievedByAccountId() {
	// GIVEN
	currentAccount, _ := ledger.NewAccount(1, "Current", "AED", ledger.MustMakeUpdatedByUserId(suite.testUser.Id()))

	// WHEN
	tx := suite.accountDao.MustBeginTx()
	_ = suite.accountDao.SaveTx(context.Background(), suite.testUser.Id(), ledger.Accounts{currentAccount}, tx)
	_ = tx.Commit()

	tx = suite.accountDao.MustBeginTx()
	account, err := suite.accountDao.GetAccountById(context.Background(), currentAccount.Id(), suite.testUser.Id(), tx)
	_ = tx.Commit()

	// THEN
	assert.Nil(suite.T(), err)
	assert.NotEqual(suite.T(), ledger.Account{}, account)

	assert.EqualValues(suite.T(), ledger.AccountId(1), account.Id())
	assert.EqualValues(suite.T(), "Current", account.Name())
	assert.EqualValues(suite.T(), "AED", account.Currency())
}

func (suite *AccountDaoTestSuite) Test_Given_twoAccounts_WHEN_theAccountsHaveTheSameName_THEN_onlyOneAccountIsSaved() {
	// GIVEN
	account1, _ := ledger.NewAccount(1, "Current", "AED", ledger.MustMakeUpdatedByUserId(suite.testUser.Id()))
	account2, _ := ledger.NewAccount(2, "Current", "AED", ledger.MustMakeUpdatedByUserId(suite.testUser.Id()))

	// WHEN
	tx := suite.accountDao.MustBeginTx()
	err1 := suite.accountDao.SaveTx(context.Background(), suite.testUser.Id(), ledger.Accounts{account1}, tx)
	_ = tx.Commit()

	tx = suite.accountDao.MustBeginTx()
	err2 := suite.accountDao.SaveTx(context.Background(), suite.testUser.Id(), ledger.Accounts{account2}, tx)
	_ = tx.Commit()

	// THEN
	assert.Nil(suite.T(), err1)
	assert.NotNil(suite.T(), err2)

	coreError := err2.(ledger.Error)
	assert.Equal(suite.T(), ledger.ErrAccountNameDuplicated, coreError.Code())
	assert.Equal(suite.T(), "Acccount named \"Current\" already exists", coreError.Error())
}

func (suite *AccountDaoTestSuite) Test_Given_twoUsersCreateTwoAccounts_WHEN_aUserTriesToRetrieveTheAccountOfTheOtherUserByAccountId_THEN_accountIsNotFound() {
	// GIVEN
	user1, _ := ledger.NewUserWithEmailString(ledger.UserId(time.Now().UnixNano()), "bob1@example.com")
	if err := suite.userDao.Save(user1); err != nil {
		log.Fatalf("AccountDaoTestSuite: Test setup failed: %s", err)
	}

	user2, _ := ledger.NewUserWithEmailString(ledger.UserId(time.Now().UnixNano()), "bob2@example.com")
	if err := suite.userDao.Save(user2); err != nil {
		log.Fatalf("AccountDaoTestSuite: Test setup failed: %s", err)
	}

	currentAccountOfUser1, _ := ledger.NewAccount(1, "Current", "AED", ledger.MustMakeUpdatedByUserId(user1.Id()))
	currentAccountOfUser2, _ := ledger.NewAccount(2, "Current", "EUR", ledger.MustMakeUpdatedByUserId(user2.Id()))

	tx := suite.accountDao.MustBeginTx()
	_ = suite.accountDao.SaveTx(context.Background(), user1.Id(), ledger.Accounts{currentAccountOfUser1}, tx)
	_ = suite.accountDao.SaveTx(context.Background(), user2.Id(), ledger.Accounts{currentAccountOfUser2}, tx)
	_ = tx.Commit()

	tx = suite.accountDao.MustBeginTx()
	account, err := suite.accountDao.GetAccountById(context.Background(), currentAccountOfUser2.Id(), user1.Id(), tx)
	_ = tx.Commit()

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ledger.Account{}, account)

	coreError := err.(ledger.Error)
	assert.Equal(suite.T(), ledger.ErrAccountNotFound, coreError.Code())
	assert.Equal(suite.T(), "Account not found", coreError.Error())
}
