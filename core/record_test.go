package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RecordTestSuite struct {
	suite.Suite
}

func TestRecordTestSuite(t *testing.T) {
	suite.Run(t, new(RecordTestSuite))
}

// -- SUITE

func (suite *RecordTestSuite) Test_GIVEN_invalidRecordId_WHEN_RecordIsCreated_THEN_errorIsReturned() {
	// GIVEN
	recordId := RecordId(0)
	category, _ := NewCategory(CategoryId(1), "Bills")
	amount, _ := NewMoney("AED", 20000)

	// WHEN
	record, err := NewRecord(recordId, "Telephone Bill", category, amount, time.Now(), Expense, 0)

	// THEN
	assert.Nil(suite.T(), record)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrRecordValidation, err.(Error).Code())
	assert.Equal(suite.T(), "Id must be greater than 0", err.(Error).Error())
	assert.Equal(suite.T(), "Id must be greater than 0", err.(Error).Fields()["id"])
}

func (suite *RecordTestSuite) Test_GIVEN_emptyNote_WHEN_RecordIsCreated_THEN_recordIsCreated() {

	// GIVEN
	recordId := RecordId(1)
	category, _ := NewCategory(CategoryId(1), "Bills")
	amount, _ := NewMoney("AED", 20000)

	// WHEN
	record, err := NewRecord(recordId, "", category, amount, time.Now(), Expense, 0)

	// THEN
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), record)
	assert.Equal(suite.T(), "", record.Note())
	assert.Equal(suite.T(), "AED -200.00", record.Amount().String())
	assert.Equal(suite.T(), "Bills", record.Category().Name())
}

func (suite *RecordTestSuite) Test_GIVEN_noteExceeds50Characters_WHEN_RecordIsCreated_THEN_errorIsReturned() {

	// GIVEN
	recordId := RecordId(1)
	category, _ := NewCategory(CategoryId(1), "Bills")
	amount, _ := NewMoney("AED", 20000)

	// WHEN
	record, err := NewRecord(recordId, "123456789012345678901234567890123456789012345678901", category, amount, time.Now(), Expense, 0)

	// THEN
	assert.Nil(suite.T(), record)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrRecordValidation, err.(Error).Code())
	assert.Equal(suite.T(), "Note can not be longer than 50 characters", err.(Error).Error())
	assert.Equal(suite.T(), "Note can not be longer than 50 characters", err.(Error).Fields()["note"])
}

func (suite *RecordTestSuite) Test_GIVEN_nilCategory_WHEN_RecordIsCreated_THEN_errorIsReturned() {

	// GIVEN
	recordId := RecordId(1)
	amount, _ := NewMoney("AED", 20000)

	// WHEN
	record, err := NewRecord(recordId, "Telephone Bill", nil, amount, time.Now(), Expense, 0)

	// THEN
	assert.Nil(suite.T(), record)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrRecordValidation, err.(Error).Code())
	assert.Equal(suite.T(), "Category is required", err.(Error).Error())
	assert.Equal(suite.T(), "Category is required", err.(Error).Fields()["category"])
}

func (suite *RecordTestSuite) Test_GIVEN_nilTime_WHEN_RecordIsCreated_THEN_errorIsReturned() {

	// GIVEN
	recordId := RecordId(1)
	category, _ := NewCategory(CategoryId(1), "Bills")
	amount, _ := NewMoney("AED", 20000)

	// WHEN
	record, err := NewRecord(recordId, "Telephone Bill", category, amount, time.Time{}, Expense, 0)

	// THEN
	assert.Nil(suite.T(), record)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrRecordValidation, err.(Error).Code())
	assert.Equal(suite.T(), "Invalid date", err.(Error).Error())
	assert.Equal(suite.T(), "Invalid date", err.(Error).Fields()["date"])
}

func (suite *RecordTestSuite) Test_GIVEN_nilAmount_WHEN_RecordIsCreated_THEN_errorIsReturned() {

	// GIVEN
	recordId := RecordId(1)
	category, _ := NewCategory(CategoryId(1), "Bills")

	// WHEN
	record, err := NewRecord(recordId, "Telephone Bill", category, nil, time.Now(), Expense, 0)

	// THEN
	assert.Nil(suite.T(), record)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrRecordValidation, err.(Error).Code())
	assert.Equal(suite.T(), "Amount is required", err.(Error).Error())
	assert.Equal(suite.T(), "Amount is required", err.(Error).Fields()["amount"])
}

func (suite *RecordTestSuite) Test_GIVEN_invalidExpenseType_WHEN_RecordIsCreated_THEN_errorIsReturned() {

	// GIVEN
	recordId := RecordId(1)
	category, _ := NewCategory(CategoryId(1), "Bills")
	amount, _ := NewMoney("AED", 20000)

	// WHEN
	record, err := NewRecord(recordId, "Telephone Bill", category, amount, time.Now().UTC(), "nil", 0)

	// THEN
	assert.Nil(suite.T(), record)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrRecordValidation, err.(Error).Code())
	assert.Equal(suite.T(), "recordType must be INCOME,EXPENSE or TRANSFER. Invalid: \"nil\"", err.(Error).Error())
	assert.Equal(suite.T(), "recordType must be INCOME,EXPENSE or TRANSFER. Invalid: \"nil\"", err.(Error).Fields()["record_type"])
}

func (suite *RecordTestSuite) Test_GIVEN_transferRecordTypeWithoutBeneficiaryId_WHEN_RecordIsCreated_THEN_errorIsReturned() {

	// GIVEN
	recordId := RecordId(1)
	category, _ := NewCategory(CategoryId(1), "Bills")
	amount, _ := NewMoney("AED", 20000)

	// WHEN
	record, err := NewRecord(recordId, "Telephone Bill", category, amount, time.Now().UTC(), Transfer, 0)

	// THEN
	assert.Nil(suite.T(), record)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrRecordValidation, err.(Error).Code())
	assert.Equal(suite.T(), "beneficiaryId can not be <= 0 when record type is TRANSFER", err.(Error).Error())
	assert.Equal(suite.T(), "beneficiaryId can not be <= 0 when record type is TRANSFER", err.(Error).Fields()["beneficiary_id"])
}

func (suite *RecordTestSuite) Test_GIVEN_transferRecordTypeWithBeneficiaryId_WHEN_RecordIsCreated_THEN_recordIsCreated() {

	// GIVEN
	recordId := RecordId(1)
	category, _ := NewCategory(CategoryId(1), "Bills")
	amount, _ := NewMoney("AED", 20000)

	// WHEN
	record, err := NewRecord(recordId, "Telephone Bill", category, amount, time.Now().UTC(), Transfer, AccountId(2))

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
	category, _ := NewCategory(CategoryId(1), "Bills")
	amount, _ := NewMoney("AED", 20000)

	// WHEN
	record, err := NewRecord(recordId, "Telephone Bill", category, amount, time.Now().UTC(), Expense, AccountId(2))

	// THEN
	assert.Nil(suite.T(), record)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrRecordValidation, err.(Error).Code())
	assert.Equal(suite.T(), "beneficiaryId must be 0 when record type is \"EXPENSE\"", err.(Error).Error())
	assert.Equal(suite.T(), "beneficiaryId must be 0 when record type is \"EXPENSE\"", err.(Error).Fields()["beneficiary_id"])
}

func (suite *RecordTestSuite) Test_GIVEN_expenseRecordTypeWithZeroAmount_WHEN_RecordIsCreated_THEN_errorReturned() {

	// GIVEN
	recordId := RecordId(1)
	category, _ := NewCategory(CategoryId(1), "Bills")
	amount, _ := NewMoney("AED", 0)

	// WHEN
	record, err := NewRecord(recordId, "Telephone Bill", category, amount, time.Now().UTC(), Expense, 0)

	// THEN
	assert.Nil(suite.T(), record)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrRecordValidation, err.(Error).Code())
	assert.Equal(suite.T(), "amount must not be zero", err.(Error).Error())
	assert.Equal(suite.T(), "amount must not be zero", err.(Error).Fields()["amount"])
}

func (suite *RecordTestSuite) Test_GIVEN_expenseRecordTypeWithPositiveAmount_WHEN_RecordIsCreated_THEN_recordCreatedWithNegativeAmount() {

	// GIVEN
	recordId := RecordId(1)
	category, _ := NewCategory(CategoryId(1), "Bills")
	amount, _ := NewMoney("AED", 20000)

	// WHEN
	record, err := NewRecord(recordId, "Telephone Bill", category, amount, time.Now().UTC(), Expense, 0)

	// THEN
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), record)
	assert.Equal(suite.T(), "AED -200.00", record.Amount().String())
}

func (suite *RecordTestSuite) Test_GIVEN_expenseRecordTypeWithNegativeAmount_WHEN_RecordIsCreated_THEN_recordCreatedWithNegativeAmount() {

	// GIVEN
	recordId := RecordId(1)
	category, _ := NewCategory(CategoryId(1), "Bills")
	amount, _ := NewMoney("AED", -20000)

	// WHEN
	record, err := NewRecord(recordId, "Telephone Bill", category, amount, time.Now().UTC(), Expense, 0)

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
	category, _ := NewCategory(CategoryId(1), "Bills")
	amount, _ := NewMoney("AED", -20000)

	// WHEN
	record, err := NewRecord(recordId, "Telephone Bill", category, amount, time.Now().UTC(), Income, 0)

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
	category, _ := NewCategory(CategoryId(1), "Bills")
	amount, _ := NewMoney("AED", 20000)

	// WHEN
	record, err := NewRecord(recordId, "Telephone Bill", category, amount, time.Now().UTC(), Income, 0)

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
	category, _ := NewCategory(CategoryId(1), "Bills")
	amount, _ := NewMoney("AED", 20000)
	date := time.Date(2021, time.July, 2, 21, 10, 0, 0, time.UTC)

	// WHEN
	record, _ := NewRecord(recordId, "Telephone Bill", category, amount, date, Expense, 0)

	// THEN
	assert.Equal(suite.T(), "Record{id: 1, type: EXPENSE, amount: AED -200.00, category: Category{id: 1, name: Bills}, date: 2021-07-02T21:10:00+0000, beneficiaryId: 0}", record.String())
}
