package application

import (
	"log"
	"testing"
	"time"

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
	testSalaryCategory  *core.Category
	testBillsCategory   *core.Category
	testSavingsCategory *core.Category
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
	testSavingsCategory, _ = core.NewCategory(testSavingsCategoryId, testSavingsCategoryName)
	if err := suite.categoryDao.Save(testUserId, core.Categories{testSalaryCategory, testBillsCategory, testSavingsCategory}); err != nil {
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

func (suite *RecordDaoTestSuite) Test_Given_anExpenseRecord_WHEN_theRecordIsSaved_THEN_recordCanBeRetrievedInMonthRange() {
	// GIVEN
	aRecord, _ := core.NewRecord(core.RecordId(1), "Electricity Bill", testBillsCategory, quickMoney("AED", 20000), testRecordDate, core.Expense, 0)

	// WHEN
	_ = suite.recordDao.Save(testCurrentAccountId, aRecord)
	records, err := suite.recordDao.GetRecordsForMonth(testCurrentAccountId, int(testRecordDate.Month()), testRecordDate.Year())

	// THEN
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 1, records.Len())
	assert.EqualValues(suite.T(), core.RecordId(1), records[0].Id())
	assert.EqualValues(suite.T(), "Electricity Bill", records[0].Note())
	assert.EqualValues(suite.T(), testBillsCategoryName, records[0].Category().Name())
	assert.EqualValues(suite.T(), "AED -200.00", records[0].Amount().String())
	assert.EqualValues(suite.T(), core.Expense, records[0].Type())
	assert.EqualValues(suite.T(), "2021-07-05T00:00:00+0000", records[0].DateUTCString())
	assert.EqualValues(suite.T(), 0, records[0].BeneficiaryId())
}

func (suite *RecordDaoTestSuite) Test_Given_aTransferRecord_WHEN_theRecordIsSaved_THEN_recordCanBeRetrievedInMonthRange() {
	// GIVEN
	aRecord, _ := core.NewRecord(core.RecordId(1), "Savings", testSavingsCategory, quickMoney("AED", 50000), testRecordDate, core.Transfer, testSavingsAccountId)

	// WHEN
	_ = suite.recordDao.Save(testCurrentAccountId, aRecord)
	records, err := suite.recordDao.GetRecordsForMonth(testCurrentAccountId, int(testRecordDate.Month()), testRecordDate.Year())

	// THEN
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 1, records.Len())
	assert.EqualValues(suite.T(), core.RecordId(1), records[0].Id())
	assert.EqualValues(suite.T(), "Savings", records[0].Note())
	assert.EqualValues(suite.T(), testSavingsCategoryName, records[0].Category().Name())
	assert.EqualValues(suite.T(), "AED -500.00", records[0].Amount().String())
	assert.EqualValues(suite.T(), core.Transfer, records[0].Type())
	assert.EqualValues(suite.T(), "2021-07-05T00:00:00+0000", records[0].DateUTCString())
	assert.EqualValues(suite.T(), core.AccountId(2), records[0].BeneficiaryId())
}

func (suite *RecordDaoTestSuite) Test_Given_records_WHEN_loadingRecordsForLastPeriod_THEN_recordsForMonthOfLatestRecordReturned() {
	// GIVEN
	beforeLastMonthIncome, _ := core.NewRecord(core.RecordId(1), "Salary", testSalaryCategory, quickMoney("AED", 100000), time.Date(2021, time.May, 6, 18, 30, 0, 0, time.UTC), core.Income, 0)
	beforeLastMonthExpense, _ := core.NewRecord(core.RecordId(2), "Bills", testBillsCategory, quickMoney("AED", 50000), time.Date(2021, time.May, 6, 18, 30, 1, 0, time.UTC), core.Expense, 0)
	lastMonthIncome, _ := core.NewRecord(core.RecordId(3), "Salary", testSalaryCategory, quickMoney("AED", 100000), time.Date(2021, time.July, 6, 18, 30, 0, 0, time.UTC), core.Income, 0)
	lastMonthExpense, _ := core.NewRecord(core.RecordId(4), "Bills", testBillsCategory, quickMoney("AED", 50000), time.Date(2021, time.July, 6, 18, 30, 1, 0, time.UTC), core.Expense, 0)

	// WHEN
	_ = suite.recordDao.Save(testCurrentAccountId, lastMonthExpense)
	_ = suite.recordDao.Save(testCurrentAccountId, lastMonthIncome)
	_ = suite.recordDao.Save(testCurrentAccountId, beforeLastMonthExpense)
	_ = suite.recordDao.Save(testCurrentAccountId, beforeLastMonthIncome)

	records, err := suite.recordDao.GetRecordsForLastPeriod(testCurrentAccountId)

	// THEN
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 2, records.Len())

	assert.EqualValues(suite.T(), core.RecordId(3), records[1].Id())
	assert.EqualValues(suite.T(), "Salary", records[1].Note())
	assert.EqualValues(suite.T(), testSalaryCategoryName, records[1].Category().Name())
	assert.EqualValues(suite.T(), "AED 1000.00", records[1].Amount().String())
	assert.EqualValues(suite.T(), core.Income, records[1].Type())
	assert.EqualValues(suite.T(), "2021-07-06T00:00:00+0000", records[1].DateUTCString())
	assert.EqualValues(suite.T(), 0, records[1].BeneficiaryId())

	assert.EqualValues(suite.T(), core.RecordId(4), records[0].Id())
	assert.EqualValues(suite.T(), "Bills", records[0].Note())
	assert.EqualValues(suite.T(), testBillsCategoryName, records[0].Category().Name())
	assert.EqualValues(suite.T(), "AED -500.00", records[0].Amount().String())
	assert.EqualValues(suite.T(), core.Expense, records[0].Type())
	assert.EqualValues(suite.T(), "2021-07-06T00:00:00+0000", records[0].DateUTCString())
	assert.EqualValues(suite.T(), 0, records[0].BeneficiaryId())
}

func (suite *RecordDaoTestSuite) Test_Given_records_WHEN_searchingBySearchTerm_THEN_recordIsFound() {
	// GIVEN
	accountIds, err := simulateRecords(TestDB, 1, int(time.June), 2021, int(time.July), 2021)
	assert.Nil(suite.T(), err)

	// WHEN
	fromDate := time.Date(2021, time.June, 10, 0, 0, 0, 0, time.UTC)
	toDate := time.Date(2021, time.July, 1, 0, 0, 0, 0, time.UTC)
	records, err := suite.recordDao.Search(accountIds[0], core.RecordSearch{
		SearchTerm: "Birthday",
		FromDate:   &fromDate,
		ToDate:     &toDate,
	})

	// THEN
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 1, records.Len())
}
