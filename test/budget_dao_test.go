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

type BudgetDaoTestSuite struct {
	suite.Suite
	budgetDao  dao.BudgetDao
	accountDao dao.AccountDao

	testUser ledger.User

	testCurrentAccount ledger.Account
	testSavingsAccount ledger.Account

	testSalaryCategory  ledger.Category
	testBillsCategory   ledger.Category
	testSavingsCategory ledger.Category
}

func TestBudgetDaoTestSuite(t *testing.T) {
	suite.Run(t, new(BudgetDaoTestSuite))
}

// -- SETUP

func (suite *BudgetDaoTestSuite) SetupTest() {
	aUser, _ := ledger.NewUserWithEmailString(1, "jack.torrence@theoverlook.com")
	if err := UserDao.Save(aUser); err != nil {
		log.Fatalf("AccountDaoTestSuite: Test setup failed: %s", err)
	}

	tx := UserDao.MustBeginTx()
	currentAccount, _ := ledger.NewAccount(ledger.AccountId(1), "Current", ledger.AccountTypeCurrent, "AED", ledger.MustMakeUpdatedByUserId(aUser.Id()))
	savingsAccount, _ := ledger.NewAccount(ledger.AccountId(2), "Savings", ledger.AccountTypeSaving, "AED", ledger.MustMakeUpdatedByUserId(aUser.Id()))

	if err := AccountDao.SaveTx(context.Background(), aUser.Id(), ledger.Accounts{currentAccount, savingsAccount}, tx); err != nil {
		log.Fatalf("RecordDaoTestSuite: Test setup failed: %s", err)
	}

	testSalaryCategory, _ := ledger.NewCategory(ledger.CategoryId(1), "Salary", ledger.MustMakeUpdatedByUserId(aUser.Id()))
	testBillsCategory, _ := ledger.NewCategory(ledger.CategoryId(2), "Bills", ledger.MustMakeUpdatedByUserId(aUser.Id()))
	testSavingsCategory, _ := ledger.NewCategory(ledger.CategoryId(3), "Savings", ledger.MustMakeUpdatedByUserId(aUser.Id()))
	if err := CategoryDao.SaveTx(context.Background(), aUser.Id(), ledger.Categories{testSalaryCategory, testBillsCategory, testSavingsCategory}, tx); err != nil {
		log.Fatalf("RecordDaoTestSuite: Test setup failed: %s", err)
	}
	_ = tx.Commit()

	suite.budgetDao = BudgetDao
	suite.accountDao = AccountDao

	suite.testCurrentAccount = currentAccount
	suite.testSavingsAccount = savingsAccount

	suite.testSalaryCategory = testSalaryCategory
	suite.testBillsCategory = testBillsCategory
	suite.testSavingsCategory = testSavingsCategory

	suite.testUser = aUser
}

func (suite *BudgetDaoTestSuite) TearDownTest() {
	if err := ClearTables(); err != nil {
		log.Fatalf("Failed to tear down AccountDaoTestSuite: %s", err)
	}
}

// -- SUITE

func (suite *BudgetDaoTestSuite) Test_Given_aBudget_WHEN_theBudgetIsSaved_THEN_BudgetCanBeRetrievedByUserId() {
	// GIVEN
	aBudget, _ := ledger.NewBudget(
		ledger.BudgetId(time.Now().UnixNano()),
		ledger.AccountIds{suite.testCurrentAccount.Id()},
		ledger.BudgetPeriodTypeMonth,
		ledger.CategoryBudgets{ledger.MustCategoryBudget(ledger.NewCategoryBudget(
			suite.testBillsCategory.Id(),
			ledger.MustMoney(ledger.NewMoney("AED", 1000_00)),
		))},
		ledger.MustMakeUpdatedByUserId(suite.testUser.Id()),
	)

	// WHEN
	tx := suite.accountDao.MustBeginTx()
	err := suite.budgetDao.Save(context.Background(), suite.testUser.Id(), aBudget, tx)
	if err != nil {
		log.Fatal(err)
	}
	_ = tx.Commit()

	tx = suite.accountDao.MustBeginTx()
	theBudget, err := suite.budgetDao.GetBudgetById(context.Background(), aBudget.Id(), suite.testUser.Id(), tx)
	_ = tx.Commit()

	// THEN
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), theBudget)

	assert.EqualValues(suite.T(), aBudget.Id(), theBudget.Id())
	assert.EqualValues(suite.T(), aBudget.PeriodType(), theBudget.PeriodType())
	assert.EqualValues(suite.T(), aBudget.AccountIds(), theBudget.AccountIds())
	assert.EqualValues(suite.T(), suite.testBillsCategory.Id(), theBudget.CategoryBudgets()[0].CategoryId())
	assert.EqualValues(suite.T(), "AED 1000.00", theBudget.CategoryBudgets()[0].MaxLimit().String())
	assert.EqualValues(suite.T(), "UserId: 1", theBudget.CreatedBy().String())
}

func (suite *AccountDaoTestSuite) Test_Given_twoUsersCreateTwoBudgets_WHEN_aUserTriesToRetrieveTheBudgetOfTheOtherUserByBudgetId_THEN_budgetIsNotFound() {
	// TODO
}
