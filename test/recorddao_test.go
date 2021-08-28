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

type RecordDaoTestSuite struct {
	suite.Suite
	recordDao dao.RecordDao

	testUser ledger.User

	testCurrentAccount ledger.Account
	testSavingsAccount ledger.Account

	testSalaryCategory  ledger.Category
	testBillsCategory   ledger.Category
	testSavingsCategory ledger.Category

	testRecordDate time.Time
}

func TestRecordDaoTestSuite(t *testing.T) {
	suite.Run(t, new(RecordDaoTestSuite))
}

// -- SETUP

func (suite *RecordDaoTestSuite) SetupTest() {
	aUser, _ := ledger.NewUserWithEmailString(1, "jack.torrence@theoverlook.com")
	if err := UserDao.Save(aUser); err != nil {
		log.Fatalf("RecordDaoTestSuite: Test setup failed: %s", err)
	}

	tx := UserDao.MustBeginTx()
	currentAccount, _ := ledger.NewAccount(ledger.AccountId(1), "Current", "AED", ledger.MustMakeUpdatedByUserId(aUser.Id()))
	savingsAccount, _ := ledger.NewAccount(ledger.AccountId(2), "Savings", "AED", ledger.MustMakeUpdatedByUserId(aUser.Id()))

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

	suite.recordDao = RecordDao

	suite.testUser = aUser

	suite.testCurrentAccount = currentAccount
	suite.testSavingsAccount = savingsAccount

	suite.testRecordDate = time.Date(2021, time.July, 5, 18, 30, 0, 0, time.UTC)

	suite.testSalaryCategory = testSalaryCategory
	suite.testBillsCategory = testBillsCategory
	suite.testSavingsCategory = testSavingsCategory
}

func (suite *RecordDaoTestSuite) TearDownTest() {
	if err := ClearTables(); err != nil {
		log.Fatalf("Failed to tear down RecordDaoTestSuite: %s", err)
	}
}

// -- SUITE

func (suite *RecordDaoTestSuite) Test_WHEN_NewRecordIdIsCalled_THEN_recordIdIsReturnedFromDatabaseSequence() {
	// WHEN
	tx, _ := suite.recordDao.BeginTx()
	recordId, err := suite.recordDao.NewRecordId(tx)
	_ = tx.Commit()

	// THEN
	assert.Nil(suite.T(), err)
	assert.Positive(suite.T(), recordId)
}

func (suite *RecordDaoTestSuite) Test_Given_anIncomeRecord_WHEN_theRecordIsSaved_THEN_recordCanBeRetrievedInMonthRange() {
	// GIVEN
	aRecord, _ := ledger.NewRecord(ledger.RecordId(1), "Salary", suite.testSalaryCategory, quickMoney("AED", 100000), suite.testRecordDate, ledger.Income, ledger.NoSourceAccount, ledger.NoBeneficiaryAccount, ledger.NoTransferReference, ledger.MustMakeUpdatedByUserId(suite.testUser.Id()))

	// WHEN
	tx, _ := suite.recordDao.BeginTx()
	_ = suite.recordDao.SaveTx(context.Background(), suite.testCurrentAccount.Id(), aRecord, tx)
	_ = tx.Commit()

	records, err := suite.recordDao.GetRecordsForMonth(suite.testCurrentAccount.Id(), ledger.MakeCalendarMonthFromDate(suite.testRecordDate))

	// THEN
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 1, records.Len())
	assert.EqualValues(suite.T(), ledger.RecordId(1), records[0].Id())
	assert.EqualValues(suite.T(), "Salary", records[0].Note())
	assert.EqualValues(suite.T(), suite.testSalaryCategory.Name(), records[0].Category().Name())
	assert.EqualValues(suite.T(), "AED 1000.00", records[0].Amount().String())
	assert.EqualValues(suite.T(), ledger.Income, records[0].Type())
	assert.EqualValues(suite.T(), "2021-07-05T00:00:00+0000", records[0].DateUTCString())
	assert.EqualValues(suite.T(), 0, records[0].BeneficiaryId())
}

func (suite *RecordDaoTestSuite) Test_Given_anExpenseRecord_WHEN_theRecordIsSaved_THEN_recordCanBeRetrievedInMonthRange() {
	// GIVEN
	aRecord, _ := ledger.NewRecord(ledger.RecordId(1), "Electricity Bill", suite.testBillsCategory, quickMoney("AED", 20000), suite.testRecordDate, ledger.Expense, ledger.NoSourceAccount, ledger.NoBeneficiaryAccount, ledger.NoTransferReference, ledger.MustMakeUpdatedByUserId(suite.testUser.Id()))

	// WHEN
	tx, _ := suite.recordDao.BeginTx()
	_ = suite.recordDao.SaveTx(context.Background(), suite.testCurrentAccount.Id(), aRecord, tx)
	_ = tx.Commit()

	records, err := suite.recordDao.GetRecordsForMonth(suite.testCurrentAccount.Id(), ledger.MakeCalendarMonthFromDate(suite.testRecordDate))

	// THEN
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 1, records.Len())
	assert.EqualValues(suite.T(), ledger.RecordId(1), records[0].Id())
	assert.EqualValues(suite.T(), "Electricity Bill", records[0].Note())
	assert.EqualValues(suite.T(), suite.testBillsCategory.Name(), records[0].Category().Name())
	assert.EqualValues(suite.T(), "AED -200.00", records[0].Amount().String())
	assert.EqualValues(suite.T(), ledger.Expense, records[0].Type())
	assert.EqualValues(suite.T(), "2021-07-05T00:00:00+0000", records[0].DateUTCString())
	assert.EqualValues(suite.T(), 0, records[0].BeneficiaryId())
}

func (suite *RecordDaoTestSuite) Test_Given_aTransferRecord_WHEN_theRecordIsSaved_THEN_recordCanBeRetrievedInMonthRange() {
	// GIVEN
	aRecord, _ := ledger.NewRecord(ledger.RecordId(1), "Savings", suite.testSavingsCategory, quickMoney("AED", -50000), suite.testRecordDate, ledger.Transfer, suite.testCurrentAccount.Id(), suite.testSavingsAccount.Id(), ledger.MakeTransferReference(), ledger.MustMakeUpdatedByUserId(suite.testUser.Id()))

	// WHEN
	tx, _ := suite.recordDao.BeginTx()
	_ = suite.recordDao.SaveTx(context.Background(), suite.testCurrentAccount.Id(), aRecord, tx)
	_ = tx.Commit()

	records, err := suite.recordDao.GetRecordsForMonth(suite.testCurrentAccount.Id(), ledger.MakeCalendarMonthFromDate(suite.testRecordDate))

	// THEN
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 1, records.Len())
	assert.EqualValues(suite.T(), ledger.RecordId(1), records[0].Id())
	assert.EqualValues(suite.T(), "Savings", records[0].Note())
	assert.EqualValues(suite.T(), suite.testSavingsCategory.Name(), records[0].Category().Name())
	assert.EqualValues(suite.T(), "AED -500.00", records[0].Amount().String())
	assert.EqualValues(suite.T(), ledger.Transfer, records[0].Type())
	assert.EqualValues(suite.T(), "2021-07-05T00:00:00+0000", records[0].DateUTCString())
	assert.EqualValues(suite.T(), ledger.AccountId(2), records[0].BeneficiaryId())
}

func (suite *RecordDaoTestSuite) Test_Given_records_WHEN_loadingRecordsForLastPeriod_THEN_recordsForMonthOfLatestRecordReturned() {
	// GIVEN
	beforeLastMonthIncome, _ := ledger.NewRecord(ledger.RecordId(1), "Salary", suite.testSalaryCategory, quickMoney("AED", 100000), time.Date(2021, time.May, 6, 18, 30, 0, 0, time.UTC), ledger.Income, ledger.NoSourceAccount, ledger.NoBeneficiaryAccount, ledger.NoTransferReference, ledger.MustMakeUpdatedByUserId(suite.testUser.Id()))
	beforeLastMonthExpense, _ := ledger.NewRecord(ledger.RecordId(2), "Bills", suite.testBillsCategory, quickMoney("AED", 50000), time.Date(2021, time.May, 6, 18, 30, 1, 0, time.UTC), ledger.Expense, ledger.NoSourceAccount, ledger.NoBeneficiaryAccount, ledger.NoTransferReference, ledger.MustMakeUpdatedByUserId(suite.testUser.Id()))
	lastMonthIncome, _ := ledger.NewRecord(ledger.RecordId(3), "Salary", suite.testSalaryCategory, quickMoney("AED", 100000), time.Date(2021, time.July, 6, 18, 30, 0, 0, time.UTC), ledger.Income, ledger.NoSourceAccount, ledger.NoBeneficiaryAccount, ledger.NoTransferReference, ledger.MustMakeUpdatedByUserId(suite.testUser.Id()))
	lastMonthExpense, _ := ledger.NewRecord(ledger.RecordId(4), "Bills", suite.testBillsCategory, quickMoney("AED", 50000), time.Date(2021, time.July, 6, 18, 30, 1, 0, time.UTC), ledger.Expense, ledger.NoSourceAccount, ledger.NoBeneficiaryAccount, ledger.NoTransferReference, ledger.MustMakeUpdatedByUserId(suite.testUser.Id()))

	// WHEN
	tx, _ := suite.recordDao.BeginTx()
	_ = suite.recordDao.SaveTx(context.Background(), suite.testCurrentAccount.Id(), lastMonthExpense, tx)
	_ = suite.recordDao.SaveTx(context.Background(), suite.testCurrentAccount.Id(), lastMonthIncome, tx)
	_ = suite.recordDao.SaveTx(context.Background(), suite.testCurrentAccount.Id(), beforeLastMonthExpense, tx)
	_ = suite.recordDao.SaveTx(context.Background(), suite.testCurrentAccount.Id(), beforeLastMonthIncome, tx)
	_ = tx.Commit()

	tx, _ = suite.recordDao.BeginTx()
	records, err := suite.recordDao.GetRecordsForLastPeriod(context.Background(), suite.testCurrentAccount.Id(), tx)
	_ = tx.Commit()

	// THEN
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 2, records.Len())

	assert.EqualValues(suite.T(), ledger.RecordId(3), records[1].Id())
	assert.EqualValues(suite.T(), "Salary", records[1].Note())
	assert.EqualValues(suite.T(), suite.testSalaryCategory.Name(), records[1].Category().Name())
	assert.EqualValues(suite.T(), "AED 1000.00", records[1].Amount().String())
	assert.EqualValues(suite.T(), ledger.Income, records[1].Type())
	assert.EqualValues(suite.T(), "2021-07-06T00:00:00+0000", records[1].DateUTCString())
	assert.EqualValues(suite.T(), 0, records[1].BeneficiaryId())

	assert.EqualValues(suite.T(), ledger.RecordId(4), records[0].Id())
	assert.EqualValues(suite.T(), "Bills", records[0].Note())
	assert.EqualValues(suite.T(), suite.testBillsCategory.Name(), records[0].Category().Name())
	assert.EqualValues(suite.T(), "AED -500.00", records[0].Amount().String())
	assert.EqualValues(suite.T(), ledger.Expense, records[0].Type())
	assert.EqualValues(suite.T(), "2021-07-06T00:00:00+0000", records[0].DateUTCString())
	assert.EqualValues(suite.T(), 0, records[0].BeneficiaryId())
}

func (suite *RecordDaoTestSuite) Test_Given_records_WHEN_searchingBySearchTerm_THEN_recordIsFound() {
	// GIVEN
	userAndAccounts, err := simulateRecords(TestDB, 1, ledger.MakeCalendarMonth(2021, time.June), ledger.MakeCalendarMonth(2021, time.July))
	assert.Nil(suite.T(), err)

	// WHEN
	userId := userAndAccounts.First()
	fromDate := time.Date(2021, time.June, 10, 0, 0, 0, 0, time.UTC)
	toDate := time.Date(2021, time.July, 1, 0, 0, 0, 0, time.UTC)
	records, err := suite.recordDao.Search(userAndAccounts[userId][0], dao.RecordSearch{
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
	aRecord, _ := ledger.NewRecord(ledger.RecordId(1), "Savings", suite.testSavingsCategory, quickMoney("USD", -50000), suite.testRecordDate, ledger.Transfer, suite.testCurrentAccount.Id(), suite.testSavingsAccount.Id(), ledger.MakeTransferReference(), ledger.MustMakeUpdatedByUserId(suite.testUser.Id()))

	// WHEN
	tx, _ := suite.recordDao.BeginTx()
	err := suite.recordDao.SaveTx(context.Background(), suite.testCurrentAccount.Id(), aRecord, tx)
	_ = tx.Commit()

	records, _ := suite.recordDao.GetRecordsForMonth(suite.testCurrentAccount.Id(), ledger.MakeCalendarMonthFromDate(suite.testRecordDate))

	// THEN
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 1, records.Len())
	assert.EqualValues(suite.T(), ledger.RecordId(1), records[0].Id())
	assert.EqualValues(suite.T(), "Savings", records[0].Note())
	assert.EqualValues(suite.T(), suite.testSavingsCategory.Name(), records[0].Category().Name())
	assert.EqualValues(suite.T(), "AED -500.00", records[0].Amount().String())
	assert.EqualValues(suite.T(), ledger.Transfer, records[0].Type())
	assert.EqualValues(suite.T(), "2021-07-05T00:00:00+0000", records[0].DateUTCString())
	assert.EqualValues(suite.T(), ledger.AccountId(2), records[0].BeneficiaryId())
}
