package application

import (
	"log"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/w-k-s/simple-budget-tracker/core"
)

type RecordDaoTestSuite struct {
	suite.Suite
	userDao     core.UserDao
	accountDao  core.AccountDao
	categoryDao core.CategoryDao
	recordDao   core.RecordDao
}

func TestRecordDaoTestSuite(t *testing.T) {
	suite.Run(t, new(RecordDaoTestSuite))
}

// -- SETUP

var (
	testSalaryCategory *core.Category
	testBillsCategory  *core.Category
)

func (suite *RecordDaoTestSuite) SetupTest() {
	suite.userDao = UserDao
	suite.accountDao = AccountDao
	suite.categoryDao = CategoryDao
	suite.recordDao = RecordDao

	aUser, _ := core.NewUserWithEmailString(testUserId, testUserEmail)
	if err := suite.userDao.Save(aUser); err != nil {
		log.Fatalf("RecordDaoTestSuite: Test setup failed: %s", err)
	}

	currentAccount, _ := core.NewAccount(testCurrentAccountId, testCurrentAccountName, testCurrentAccountCurrency)
	if err := suite.accountDao.Save(testUserId, currentAccount); err != nil {
		log.Fatalf("RecordDaoTestSuite: Test setup failed: %s", err)
	}

	savingsAccount, _ := core.NewAccount(testSavingsAccountId, testSavingsAccountName, testSavingsAccountCurrency)
	if err := suite.accountDao.Save(testUserId, savingsAccount); err != nil {
		log.Fatalf("RecordDaoTestSuite: Test setup failed: %s", err)
	}

	testSalaryCategory, _ = core.NewCategory(testSalaryCategoryId, testSalaryCategoryName)
	testBillsCategory, _ = core.NewCategory(testBillsCategoryId, testBillsCategoryName)
	if err := suite.categoryDao.Save(testUserId, core.Categories{testSalaryCategory, testBillsCategory}); err != nil {
		log.Fatalf("RecordDaoTestSuite: Test setup failed: %s", err)
	}
}

func (suite *RecordDaoTestSuite) TearDownTest() {
	if err := ClearTables(); err != nil {
		log.Fatalf("Failed to tear down RecordDaoTestSuite: %s", err)
	}
}

// -- SUITE

func (suite *RecordDaoTestSuite) Test_WHEN_NewRecordIdIsCalled_THEN_recordIdIsReturnedFromDatabaseSequence() {
	// WHEN
	recordId, err := suite.recordDao.NewRecordId()

	// THEN
	assert.Nil(suite.T(), err)
	assert.Positive(suite.T(), recordId)
}

func (suite *RecordDaoTestSuite) Test_Given_anIncomeRecord_WHEN_theRecordIsSaved_THEN_recordCanBeRetrievedInMonthRange() {
	// GIVEN
	aRecord, _ := core.NewRecord(core.RecordId(1), "Salary", testSalaryCategory, quickMoney("AED", 100000), testRecordDate, core.Income, 0)

	// WHEN
	_ = suite.recordDao.Save(testCurrentAccountId, aRecord)
	records, err := suite.recordDao.GetRecordsForMonth(testCurrentAccountId, int(testRecordDate.Month()), testRecordDate.Year())

	// THEN
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 1, records.Len())
	assert.EqualValues(suite.T(), core.RecordId(1), records[0].Id())
	assert.EqualValues(suite.T(), "Salary", records[0].Note())
	assert.EqualValues(suite.T(), testSalaryCategoryName, records[0].Category().Name())
	assert.EqualValues(suite.T(), "AED 1000.00", records[0].Amount().String())
	assert.EqualValues(suite.T(), core.Income, records[0].Type())
	assert.EqualValues(suite.T(), "2021-07-05T00:00:00+0000", records[0].DateUTCString())
	assert.EqualValues(suite.T(), 0, records[0].BeneficiaryId())
}
