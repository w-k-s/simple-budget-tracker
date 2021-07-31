package ledger

import (
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RecordTestSuite struct {
	suite.Suite
	billsCategory Category
	billAmount    Money
}

func TestRecordTestSuite(t *testing.T) {
	suite.Run(t, new(RecordTestSuite))
}

func (suite *RecordTestSuite) SetupTest() {
	category, _ := NewCategory(CategoryId(1), "Bills", MustMakeUpdatedByUserId(UserId(1)))
	amount, _ := NewMoney("AED", 20000)

	suite.billsCategory = category
	suite.billAmount = amount
}

// -- SUITE

func (suite *RecordTestSuite) Test_GIVEN_invalidRecordId_WHEN_RecordIsCreated_THEN_errorIsReturned() {
	// GIVEN
	recordId := RecordId(0)

	// WHEN
	record, err := NewRecord(recordId, "Telephone Bill", suite.billsCategory, suite.billAmount, time.Now(), Expense, 0, MustMakeUpdatedByUserId(UserId(1)))

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), Record{}, record)
	assert.Equal(suite.T(), ErrRecordValidation, err.(Error).Code())
	assert.Equal(suite.T(), "Id must be greater than 0", err.(Error).Error())
	assert.Equal(suite.T(), "Id must be greater than 0", err.(Error).Fields()["id"])
}

func (suite *RecordTestSuite) Test_GIVEN_emptyNote_WHEN_RecordIsCreated_THEN_recordIsCreated() {

	// GIVEN
	recordId := RecordId(1)

	// WHEN
	record, err := NewRecord(recordId, "", suite.billsCategory, suite.billAmount, time.Now(), Expense, 0, MustMakeUpdatedByUserId(UserId(1)))

	// THEN
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), record)
	assert.Equal(suite.T(), "", record.Note())
	assert.Equal(suite.T(), "AED -200.00", record.Amount().String())
	assert.Equal(suite.T(), "Bills", record.Category().Name())
	assert.Equal(suite.T(), "UserId: 1", record.CreatedBy().String())
	assert.Equal(suite.T(), Version(1), record.Version())
	assert.True(suite.T(), time.Now().UTC().Sub(record.CreatedAtUTC()) < time.Duration(1)*time.Second)
}

func (suite *RecordTestSuite) Test_GIVEN_noteExceeds50Characters_WHEN_RecordIsCreated_THEN_errorIsReturned() {

	// GIVEN
	recordId := RecordId(1)

	// WHEN
	record, err := NewRecord(recordId, "123456789012345678901234567890123456789012345678901", suite.billsCategory, suite.billAmount, time.Now(), Expense, 0, MustMakeUpdatedByUserId(UserId(1)))

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), Record{}, record)
	assert.Equal(suite.T(), ErrRecordValidation, err.(Error).Code())
	assert.Equal(suite.T(), "Note can not be longer than 50 characters", err.(Error).Error())
	assert.Equal(suite.T(), "Note can not be longer than 50 characters", err.(Error).Fields()["note"])
}

func (suite *RecordTestSuite) Test_GIVEN_nilCategory_WHEN_RecordIsCreated_THEN_errorIsReturned() {

	// GIVEN
	recordId := RecordId(1)

	// WHEN
	record, err := NewRecord(recordId, "Telephone Bill", Category{}, suite.billAmount, time.Now(), Expense, 0, MustMakeUpdatedByUserId(UserId(1)))

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), Record{}, record)
	assert.Equal(suite.T(), ErrRecordValidation, err.(Error).Code())
	assert.Equal(suite.T(), "Category is required", err.(Error).Error())
	assert.Equal(suite.T(), "Category is required", err.(Error).Fields()["category"])
}

func (suite *RecordTestSuite) Test_GIVEN_nilTime_WHEN_RecordIsCreated_THEN_errorIsReturned() {

	// GIVEN
	recordId := RecordId(1)

	// WHEN
	record, err := NewRecord(recordId, "Telephone Bill", suite.billsCategory, suite.billAmount, time.Time{}, Expense, 0, MustMakeUpdatedByUserId(UserId(1)))

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), Record{}, record)
	assert.Equal(suite.T(), ErrRecordValidation, err.(Error).Code())
	assert.Equal(suite.T(), "Invalid date", err.(Error).Error())
	assert.Equal(suite.T(), "Invalid date", err.(Error).Fields()["date"])
}

func (suite *RecordTestSuite) Test_GIVEN_nilAmount_WHEN_RecordIsCreated_THEN_errorIsReturned() {

	// GIVEN
	recordId := RecordId(1)

	// WHEN
	record, err := NewRecord(recordId, "Telephone Bill", suite.billsCategory, nil, time.Now(), Expense, 0, MustMakeUpdatedByUserId(UserId(1)))

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), Record{}, record)
	assert.Equal(suite.T(), ErrRecordValidation, err.(Error).Code())
	assert.Equal(suite.T(), "Amount is required", err.(Error).Error())
	assert.Equal(suite.T(), "Amount is required", err.(Error).Fields()["amount"])
}

func (suite *RecordTestSuite) Test_GIVEN_invalidExpenseType_WHEN_RecordIsCreated_THEN_errorIsReturned() {

	// GIVEN
	recordId := RecordId(1)
	// WHEN
	record, err := NewRecord(recordId, "Telephone Bill", suite.billsCategory, suite.billAmount, time.Now().UTC(), "nil", 0, MustMakeUpdatedByUserId(UserId(1)))

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), Record{}, record)
	assert.Equal(suite.T(), ErrRecordValidation, err.(Error).Code())
	assert.Equal(suite.T(), "recordType must be INCOME,EXPENSE or TRANSFER. Invalid: \"nil\"", err.(Error).Error())
	assert.Equal(suite.T(), "recordType must be INCOME,EXPENSE or TRANSFER. Invalid: \"nil\"", err.(Error).Fields()["record_type"])
}

func (suite *RecordTestSuite) Test_GIVEN_transferRecordTypeWithoutBeneficiaryId_WHEN_RecordIsCreated_THEN_errorIsReturned() {

	// GIVEN
	recordId := RecordId(1)

	// WHEN
	record, err := NewRecord(recordId, "Telephone Bill", suite.billsCategory, suite.billAmount, time.Now().UTC(), Transfer, 0, MustMakeUpdatedByUserId(UserId(1)))

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), Record{}, record)
	assert.Equal(suite.T(), ErrRecordValidation, err.(Error).Code())
	assert.Equal(suite.T(), "beneficiaryId can not be <= 0 when record type is TRANSFER", err.(Error).Error())
	assert.Equal(suite.T(), "beneficiaryId can not be <= 0 when record type is TRANSFER", err.(Error).Fields()["beneficiary_id"])
}

func (suite *RecordTestSuite) Test_GIVEN_transferRecordTypeWithBeneficiaryId_WHEN_RecordIsCreated_THEN_recordIsCreated() {

	// GIVEN
	recordId := RecordId(1)

	// WHEN
	record, err := NewRecord(recordId, "Telephone Bill", suite.billsCategory, suite.billAmount, time.Now().UTC(), Transfer, AccountId(2), MustMakeUpdatedByUserId(1))

	// THEN
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), record)
	assert.Equal(suite.T(), RecordId(1), record.Id())
	assert.Equal(suite.T(), "AED -200.00", record.amount.String())
	assert.Equal(suite.T(), "Telephone Bill", record.Note())
	assert.Equal(suite.T(), "Category{id: 1, name: Bills}", record.Category().String())
	assert.Equal(suite.T(), Transfer, record.Type())
	assert.Equal(suite.T(), AccountId(2), record.BeneficiaryId())
}

func (suite *RecordTestSuite) Test_GIVEN_expenseRecordTypeWithBeneficiaryId_WHEN_RecordIsCreated_THEN_errorIsReturned() {

	// GIVEN
	recordId := RecordId(1)

	// WHEN
	record, err := NewRecord(recordId, "Telephone Bill", suite.billsCategory, suite.billAmount, time.Now().UTC(), Expense, AccountId(2), MustMakeUpdatedByUserId(1))

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), Record{}, record)
	assert.Equal(suite.T(), ErrRecordValidation, err.(Error).Code())
	assert.Equal(suite.T(), "beneficiaryId must be 0 when record type is \"EXPENSE\"", err.(Error).Error())
	assert.Equal(suite.T(), "beneficiaryId must be 0 when record type is \"EXPENSE\"", err.(Error).Fields()["beneficiary_id"])
}

func (suite *RecordTestSuite) Test_GIVEN_expenseRecordTypeWithZeroAmount_WHEN_RecordIsCreated_THEN_errorReturned() {

	// GIVEN
	recordId := RecordId(1)
	amount, _ := NewMoney("AED", 0)

	// WHEN
	record, err := NewRecord(recordId, "Telephone Bill", suite.billsCategory, amount, time.Now().UTC(), Expense, 0, MustMakeUpdatedByUserId(1))

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), Record{}, record)
	assert.Equal(suite.T(), ErrRecordValidation, err.(Error).Code())
	assert.Equal(suite.T(), "amount must not be zero", err.(Error).Error())
	assert.Equal(suite.T(), "amount must not be zero", err.(Error).Fields()["amount"])
}

func (suite *RecordTestSuite) Test_GIVEN_expenseRecordTypeWithPositiveAmount_WHEN_RecordIsCreated_THEN_recordCreatedWithNegativeAmount() {

	// GIVEN
	recordId := RecordId(1)

	// WHEN
	record, err := NewRecord(recordId, "Telephone Bill", suite.billsCategory, suite.billAmount, time.Now().UTC(), Expense, 0, MustMakeUpdatedByUserId(1))

	// THEN
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), record)
	assert.Equal(suite.T(), "AED -200.00", record.Amount().String())
}

func (suite *RecordTestSuite) Test_GIVEN_expenseRecordTypeWithNegativeAmount_WHEN_RecordIsCreated_THEN_recordCreatedWithNegativeAmount() {

	// GIVEN
	recordId := RecordId(1)

	// WHEN
	record, err := NewRecord(recordId, "Telephone Bill", suite.billsCategory, suite.billAmount, time.Now().UTC(), Expense, 0, MustMakeUpdatedByUserId(1))

	// THEN
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), record)
	assert.Equal(suite.T(), "AED -200.00", record.Amount().String())
	assert.Equal(suite.T(), "Telephone Bill", record.Note())
	assert.Equal(suite.T(), "Category{id: 1, name: Bills}", record.Category().String())
	assert.Equal(suite.T(), Expense, record.Type())
	assert.Equal(suite.T(), AccountId(0), record.BeneficiaryId())
}

func (suite *RecordTestSuite) Test_GIVEN_incomeRecordTypeWithNegativeAmount_WHEN_RecordIsCreated_THEN_recordCreatedWithPositiveAmount() {

	// GIVEN
	recordId := RecordId(1)

	// WHEN
	record, err := NewRecord(recordId, "Telephone Bill", suite.billsCategory, suite.billAmount, time.Now().UTC(), Income, 0, MustMakeUpdatedByUserId(1))

	// THEN
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), record)
	assert.Equal(suite.T(), "AED 200.00", record.Amount().String())
	assert.Equal(suite.T(), "Telephone Bill", record.Note())
	assert.Equal(suite.T(), "Category{id: 1, name: Bills}", record.Category().String())
	assert.Equal(suite.T(), Income, record.Type())
	assert.Equal(suite.T(), AccountId(0), record.BeneficiaryId())
}

func (suite *RecordTestSuite) Test_GIVEN_incomeRecordTypeWithPositiveAmount_WHEN_RecordIsCreated_THEN_recordCreatedWithPositiveAmount() {

	// GIVEN
	recordId := RecordId(1)

	// WHEN
	record, err := NewRecord(recordId, "Telephone Bill", suite.billsCategory, suite.billAmount, time.Now().UTC(), Income, 0, MustMakeUpdatedByUserId(1))

	// THEN
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), record)
	assert.Equal(suite.T(), "AED 200.00", record.Amount().String())
	assert.Equal(suite.T(), "Telephone Bill", record.Note())
	assert.Equal(suite.T(), "Category{id: 1, name: Bills}", record.Category().String())
	assert.Equal(suite.T(), Income, record.Type())
	assert.Equal(suite.T(), AccountId(0), record.BeneficiaryId())
}

func (suite *RecordTestSuite) Test_GIVEN_aRecord_WHEN_stringIsCalled_THEN_stringIsReadable() {

	// GIVEN
	recordId := RecordId(1)
	date := time.Date(2021, time.July, 2, 21, 10, 0, 0, time.UTC)

	// WHEN
	record, _ := NewRecord(recordId, "Telephone Bill", suite.billsCategory, suite.billAmount, date, Expense, 0, MustMakeUpdatedByUserId(1))

	// THEN
	assert.Equal(suite.T(), "Record{id: 1, type: EXPENSE, amount: AED -200.00, category: Category{id: 1, name: Bills}, date: 2021-07-02T21:10:00+0000, beneficiaryId: 0}", record.String())
}

func (suite *RecordTestSuite) Test_GIVEN_records_WHEN_stringIsCalled_THEN_recordsArePrintedInSortedOrder() {
	// GIVEN
	salaryCategory, _ := NewCategory(CategoryId(1), "Salary", MustMakeUpdatedByUserId(1))
	salaryAmount, _ := NewMoney("AED", 100000)
	salaryDate := time.Date(2021, time.July, 1, 12, 0, 0, 0, time.UTC)

	billsCategory, _ := NewCategory(CategoryId(2), "Bills", MustMakeUpdatedByUserId(1))
	billAmount, _ := NewMoney("AED", 5000)
	billDate := time.Date(2021, time.July, 3, 13, 30, 0, 0, time.UTC)

	record1, _ := NewRecord(RecordId(1), "Salary", salaryCategory, salaryAmount, salaryDate, Income, 0, MustMakeUpdatedByUserId(1))
	record2, _ := NewRecord(RecordId(2), "Electricity Bill", billsCategory, billAmount, billDate, Expense, 0, MustMakeUpdatedByUserId(1))

	// WHEN
	records := Records{record1, record2}

	// THEN
	assert.Equal(suite.T(), "Records{Record{id: 1, type: INCOME, amount: AED 1000.00, category: Category{id: 1, name: Salary}, date: 2021-07-01T12:00:00+0000, beneficiaryId: 0}, Record{id: 2, type: EXPENSE, amount: AED -50.00, category: Category{id: 2, name: Bills}, date: 2021-07-03T13:30:00+0000, beneficiaryId: 0}}", records.String())
}

func (suite *RecordTestSuite) Test_GIVEN_records_WHEN_recordsAreSorted_THEN_recordsAreSortedByDate() {
	// GIVEN
	salaryCategory, _ := NewCategory(CategoryId(1), "Salary", MustMakeUpdatedByUserId(1))
	salaryAmount, _ := NewMoney("AED", 100000)
	salaryDate := time.Date(2021, time.July, 1, 12, 0, 0, 0, time.UTC)

	billsCategory, _ := NewCategory(CategoryId(2), "Bills", MustMakeUpdatedByUserId(1))
	billAmount, _ := NewMoney("AED", 5000)
	billDate := time.Date(2021, time.July, 3, 13, 30, 0, 0, time.UTC)

	record1, _ := NewRecord(RecordId(1), "Salary", salaryCategory, salaryAmount, salaryDate, Income, 0, MustMakeUpdatedByUserId(1))
	record2, _ := NewRecord(RecordId(2), "Electricity Bill", billsCategory, billAmount, billDate, Expense, 0, MustMakeUpdatedByUserId(1))

	// WHEN
	records := Records{record1, record2}
	sort.Sort(records)

	// THEN
	assert.Equal(suite.T(), 2, records.Len())
	assert.Equal(suite.T(), RecordId(1), records[0].Id())
	assert.Equal(suite.T(), RecordId(2), records[1].Id())
}

func (suite *RecordTestSuite) Test_GIVEN_recordsOfSameCurrency_WHEN_recordsAreTotaled_THEN_totalIsCorrect() {
	// GIVEN
	salaryCategory, _ := NewCategory(CategoryId(1), "Salary", MustMakeUpdatedByUserId(1))
	salaryAmount, _ := NewMoney("AED", 100000)
	salaryDate := time.Date(2021, time.July, 1, 12, 0, 0, 0, time.UTC)

	billsCategory, _ := NewCategory(CategoryId(2), "Bills", MustMakeUpdatedByUserId(1))
	billAmount, _ := NewMoney("AED", 5000)
	billDate := time.Date(2021, time.July, 3, 13, 30, 0, 0, time.UTC)

	record1, _ := NewRecord(RecordId(1), "Salary", salaryCategory, salaryAmount, salaryDate, Income, 0, MustMakeUpdatedByUserId(1))
	record2, _ := NewRecord(RecordId(2), "Electricity Bill", billsCategory, billAmount, billDate, Expense, 0, MustMakeUpdatedByUserId(1))

	// WHEN
	records := Records{record1, record2}
	total, err := records.Total()

	// THEN
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), total)
	assert.Equal(suite.T(), "AED 950.00", total.String())
}

func (suite *RecordTestSuite) Test_GIVEN_recordsOfDifferentCurrency_WHEN_recordsAreTotaled_THEN_errorIsReturned() {
	// GIVEN
	salaryCategory, _ := NewCategory(CategoryId(1), "Salary", MustMakeUpdatedByUserId(1))
	salaryAmount, _ := NewMoney("AED", 100000)
	salaryDate := time.Date(2021, time.July, 1, 12, 0, 0, 0, time.UTC)

	billsCategory, _ := NewCategory(CategoryId(2), "Bills", MustMakeUpdatedByUserId(1))
	billAmount, _ := NewMoney("KWD", 5000)
	billDate := time.Date(2021, time.July, 3, 13, 30, 0, 0, time.UTC)

	record1, _ := NewRecord(RecordId(1), "Salary", salaryCategory, salaryAmount, salaryDate, Income, 0, MustMakeUpdatedByUserId(1))
	record2, _ := NewRecord(RecordId(2), "Electricity Bill", billsCategory, billAmount, billDate, Expense, 0, MustMakeUpdatedByUserId(1))

	// WHEN
	total, err := Records{record1, record2}.Total()

	// THEN
	assert.Nil(suite.T(), total)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrAmountMismatchingCurrencies, err.(Error).Code())
	assert.Equal(suite.T(), "Can not sum mismatching currencies", err.(Error).Error())
}

func (suite *RecordTestSuite) Test_GIVEN_emptyRecords_WHEN_recordsAreTotaled_THEN_errorIsReturned() {

	// WHEN
	total, err := Records{}.Total()

	// THEN
	assert.Nil(suite.T(), total)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrAmountTotalOfEmptySet, err.(Error).Code())
	assert.Equal(suite.T(), "No amounts to total", err.(Error).Error())
}
