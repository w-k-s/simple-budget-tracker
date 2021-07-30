package test

import (
	"log"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
	dao "github.com/w-k-s/simple-budget-tracker/pkg/persistence"
)

type RecordDaoTestSuite struct {
	suite.Suite
	userDao     dao.UserDao
	accountDao  dao.AccountDao
	categoryDao dao.CategoryDao
	recordDao   dao.RecordDao
}

func TestRecordDaoTestSuite(t *testing.T) {
	suite.Run(t, new(RecordDaoTestSuite))
}

// -- SETUP

var (
	testSalaryCategory  ledger.Category
	testBillsCategory   ledger.Category
	testSavingsCategory ledger.Category
)

func (suite *RecordDaoTestSuite) SetupTest() {
	suite.userDao = UserDao
	suite.accountDao = AccountDao
	suite.categoryDao = CategoryDao
	suite.recordDao = RecordDao

	aUser, _ := ledger.NewUserWithEmailString(testUserId, testUserEmail)
	if err := suite.userDao.Save(aUser); err != nil {
		log.Fatalf("RecordDaoTestSuite: Test setup failed: %s", err)
	}

	currentAccount, _ := ledger.NewAccount(testCurrentAccountId, testUserId, testCurrentAccountName, testCurrentAccountCurrency)
	if err := suite.accountDao.Save(testUserId, &currentAccount); err != nil {
		log.Fatalf("RecordDaoTestSuite: Test setup failed: %s", err)
	}

	savingsAccount, _ := ledger.NewAccount(testSavingsAccountId, testUserId, testSavingsAccountName, testSavingsAccountCurrency)
	if err := suite.accountDao.Save(testUserId, &savingsAccount); err != nil {
		log.Fatalf("RecordDaoTestSuite: Test setup failed: %s", err)
	}

	testSalaryCategory, _ = ledger.NewCategory(testSalaryCategoryId, testUserId, testSalaryCategoryName)
	testBillsCategory, _ = ledger.NewCategory(testBillsCategoryId, testUserId, testBillsCategoryName)
	testSavingsCategory, _ = ledger.NewCategory(testSavingsCategoryId, testUserId, testSavingsCategoryName)
	if err := suite.categoryDao.Save(testUserId, ledger.Categories{testSalaryCategory, testBillsCategory, testSavingsCategory}); err != nil {
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
	aRecord, _ := ledger.NewRecord(ledger.RecordId(1), testUserId, "Salary", testSalaryCategory, quickMoney("AED", 100000), testRecordDate, ledger.Income, 0)

	// WHEN
	_ = suite.recordDao.Save(testCurrentAccountId, &aRecord)
	records, err := suite.recordDao.GetRecordsForMonth(testCurrentAccountId, ledger.MakeCalendarMonthFromDate(testRecordDate))

	// THEN
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 1, records.Len())
	assert.EqualValues(suite.T(), ledger.RecordId(1), records[0].Id())
	assert.EqualValues(suite.T(), "Salary", records[0].Note())
	assert.EqualValues(suite.T(), testSalaryCategoryName, records[0].Category().Name())
	assert.EqualValues(suite.T(), "AED 1000.00", records[0].Amount().String())
	assert.EqualValues(suite.T(), ledger.Income, records[0].Type())
	assert.EqualValues(suite.T(), "2021-07-05T00:00:00+0000", records[0].DateUTCString())
	assert.EqualValues(suite.T(), 0, records[0].BeneficiaryId())
}

func (suite *RecordDaoTestSuite) Test_Given_anExpenseRecord_WHEN_theRecordIsSaved_THEN_recordCanBeRetrievedInMonthRange() {
	// GIVEN
	aRecord, _ := ledger.NewRecord(ledger.RecordId(1), testUserId, "Electricity Bill", testBillsCategory, quickMoney("AED", 20000), testRecordDate, ledger.Expense, 0)

	// WHEN
	_ = suite.recordDao.Save(testCurrentAccountId, &aRecord)
	records, err := suite.recordDao.GetRecordsForMonth(testCurrentAccountId, ledger.MakeCalendarMonthFromDate(testRecordDate))

	// THEN
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 1, records.Len())
	assert.EqualValues(suite.T(), ledger.RecordId(1), records[0].Id())
	assert.EqualValues(suite.T(), "Electricity Bill", records[0].Note())
	assert.EqualValues(suite.T(), testBillsCategoryName, records[0].Category().Name())
	assert.EqualValues(suite.T(), "AED -200.00", records[0].Amount().String())
	assert.EqualValues(suite.T(), ledger.Expense, records[0].Type())
	assert.EqualValues(suite.T(), "2021-07-05T00:00:00+0000", records[0].DateUTCString())
	assert.EqualValues(suite.T(), 0, records[0].BeneficiaryId())
}

func (suite *RecordDaoTestSuite) Test_Given_aTransferRecord_WHEN_theRecordIsSaved_THEN_recordCanBeRetrievedInMonthRange() {
	// GIVEN
	aRecord, _ := ledger.NewRecord(ledger.RecordId(1), testUserId, "Savings", testSavingsCategory, quickMoney("AED", 50000), testRecordDate, ledger.Transfer, testSavingsAccountId)

	// WHEN
	_ = suite.recordDao.Save(testCurrentAccountId, &aRecord)
	records, err := suite.recordDao.GetRecordsForMonth(testCurrentAccountId, ledger.MakeCalendarMonthFromDate(testRecordDate))

	// THEN
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 1, records.Len())
	assert.EqualValues(suite.T(), ledger.RecordId(1), records[0].Id())
	assert.EqualValues(suite.T(), "Savings", records[0].Note())
	assert.EqualValues(suite.T(), testSavingsCategoryName, records[0].Category().Name())
	assert.EqualValues(suite.T(), "AED -500.00", records[0].Amount().String())
	assert.EqualValues(suite.T(), ledger.Transfer, records[0].Type())
	assert.EqualValues(suite.T(), "2021-07-05T00:00:00+0000", records[0].DateUTCString())
	assert.EqualValues(suite.T(), ledger.AccountId(2), records[0].BeneficiaryId())
}

func (suite *RecordDaoTestSuite) Test_Given_records_WHEN_loadingRecordsForLastPeriod_THEN_recordsForMonthOfLatestRecordReturned() {
	// GIVEN
	beforeLastMonthIncome, _ := ledger.NewRecord(ledger.RecordId(1), testUserId, "Salary", testSalaryCategory, quickMoney("AED", 100000), time.Date(2021, time.May, 6, 18, 30, 0, 0, time.UTC), ledger.Income, 0)
	beforeLastMonthExpense, _ := ledger.NewRecord(ledger.RecordId(2), testUserId, "Bills", testBillsCategory, quickMoney("AED", 50000), time.Date(2021, time.May, 6, 18, 30, 1, 0, time.UTC), ledger.Expense, 0)
	lastMonthIncome, _ := ledger.NewRecord(ledger.RecordId(3), testUserId, "Salary", testSalaryCategory, quickMoney("AED", 100000), time.Date(2021, time.July, 6, 18, 30, 0, 0, time.UTC), ledger.Income, 0)
	lastMonthExpense, _ := ledger.NewRecord(ledger.RecordId(4), testUserId, "Bills", testBillsCategory, quickMoney("AED", 50000), time.Date(2021, time.July, 6, 18, 30, 1, 0, time.UTC), ledger.Expense, 0)

	// WHEN
	_ = suite.recordDao.Save(testCurrentAccountId, &lastMonthExpense)
	_ = suite.recordDao.Save(testCurrentAccountId, &lastMonthIncome)
	_ = suite.recordDao.Save(testCurrentAccountId, &beforeLastMonthExpense)
	_ = suite.recordDao.Save(testCurrentAccountId, &beforeLastMonthIncome)

	records, err := suite.recordDao.GetRecordsForLastPeriod(testCurrentAccountId)

	// THEN
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 2, records.Len())

	assert.EqualValues(suite.T(), ledger.RecordId(3), records[1].Id())
	assert.EqualValues(suite.T(), "Salary", records[1].Note())
	assert.EqualValues(suite.T(), testSalaryCategoryName, records[1].Category().Name())
	assert.EqualValues(suite.T(), "AED 1000.00", records[1].Amount().String())
	assert.EqualValues(suite.T(), ledger.Income, records[1].Type())
	assert.EqualValues(suite.T(), "2021-07-06T00:00:00+0000", records[1].DateUTCString())
	assert.EqualValues(suite.T(), 0, records[1].BeneficiaryId())

	assert.EqualValues(suite.T(), ledger.RecordId(4), records[0].Id())
	assert.EqualValues(suite.T(), "Bills", records[0].Note())
	assert.EqualValues(suite.T(), testBillsCategoryName, records[0].Category().Name())
	assert.EqualValues(suite.T(), "AED -500.00", records[0].Amount().String())
	assert.EqualValues(suite.T(), ledger.Expense, records[0].Type())
	assert.EqualValues(suite.T(), "2021-07-06T00:00:00+0000", records[0].DateUTCString())
	assert.EqualValues(suite.T(), 0, records[0].BeneficiaryId())
}

func (suite *RecordDaoTestSuite) Test_Given_records_WHEN_searchingBySearchTerm_THEN_recordIsFound() {
	// GIVEN
	accountIds, err := simulateRecords(TestDB, 1, ledger.MakeCalendarMonth(2021, time.June), ledger.MakeCalendarMonth(2021, time.July))
	assert.Nil(suite.T(), err)

	// WHEN
	fromDate := time.Date(2021, time.June, 10, 0, 0, 0, 0, time.UTC)
	toDate := time.Date(2021, time.July, 1, 0, 0, 0, 0, time.UTC)
	records, err := suite.recordDao.Search(accountIds[0], dao.RecordSearch{
		SearchTerm: "Birthday",
		FromDate:   &fromDate,
		ToDate:     &toDate,
	})

	// THEN
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 1, records.Len())
}

func (suite *RecordDaoTestSuite) Test_Given_aRecordWithAmountInDifferentCurrencyThanAccount_WHEN_recordIsSaved_THEN_currencyIsSetToAccountsCurrency() {
	// GIVEN
	aRecord, _ := ledger.NewRecord(ledger.RecordId(1), testUserId, "Savings", testSavingsCategory, quickMoney("USD", 50000), testRecordDate, ledger.Transfer, testSavingsAccountId)

	// WHEN
	err := suite.recordDao.Save(testCurrentAccountId, &aRecord)
	records, _ := suite.recordDao.GetRecordsForMonth(testCurrentAccountId, ledger.MakeCalendarMonthFromDate(testRecordDate))

	// THEN
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 1, records.Len())
	assert.EqualValues(suite.T(), ledger.RecordId(1), records[0].Id())
	assert.EqualValues(suite.T(), "Savings", records[0].Note())
	assert.EqualValues(suite.T(), testSavingsCategoryName, records[0].Category().Name())
	assert.EqualValues(suite.T(), "AED -500.00", records[0].Amount().String())
	assert.EqualValues(suite.T(), ledger.Transfer, records[0].Type())
	assert.EqualValues(suite.T(), "2021-07-05T00:00:00+0000", records[0].DateUTCString())
	assert.EqualValues(suite.T(), ledger.AccountId(2), records[0].BeneficiaryId())
}
