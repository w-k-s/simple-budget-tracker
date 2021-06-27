package core

import (
	"fmt"
	"strings"
	"time"

	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type RecordType string
const (
	Income RecordType = "INCOME"
	Expense RecordType = "EXPENSE"
	Transfer RecordType = "TRANSFER"
)

type RecordId uint64
type Record struct {
	id       RecordId
	note     string
	category     *Category
	amount		Money
	date        time.Time
	recordType  RecordType
	beneficiaryId AccountId
}

func NewRecord(id RecordId, note string, category *Category, amount Money, dateUTC time.Time, recordType RecordType, beneficiaryId AccountId) (*Record, error) {
	record := &Record{
		id:       id,
		note:     note,
		category: category,
		amount: amount,
		date: dateUTC,
		recordType: recordType,
		beneficiaryId: beneficiaryId,
	}

	errors := validate.Validate(
		&validators.IntIsGreaterThan{Name: "Id", Field: int(record.id), Compared: 0, Message: "Id must be greater than 0"},
		&validators.StringLengthInRange{Name: "Note", Field: record.note, Min: 0, Max: 50, Message: "Note can not be longer than 50 characters"},
		&NotNilValidator{Field: "Category", Value: record.category},
		&NotNilValidator{Field: "Amount", Value: record.amount},
		&validators.TimeIsPresent{Name: "Date", Field: record.date, Message: "Invalid date: %q"},
		&validators.FuncValidator{Name: "RecordType", Field: string(record.recordType), Message: "recordType must be INCOME,EXPENSE or TRANSFER. Invalid: %q", Fn: func() bool { 
			for _, rt := range []RecordType{Income,Expense,Transfer}{
				if(record.recordType == rt){
					return true
				}
			}
			return false
		}},
		&BeneficiaryIdValidator{Field: "BeneficiaryId", Value: record.beneficiaryId, RecordType: record.recordType},
	)

	if errors.HasAny() {
		flatErrors := map[string]string{}
		for field, violations := range errors.Errors {
			flatErrors[field] = strings.Join(violations, ", ")
		}
		listErrors := []string{}
		for _, violations := range flatErrors {
			listErrors = append(listErrors, violations)
		}
		return nil, NewErrorWithFields(ErrRecordValidation, strings.Join(listErrors, ", "), errors, flatErrors)
	}
	return record, nil
}

func (r Record) Id() RecordId {
	return r.id
}

func (r Record) Note() string {
	return r.note
}

func (r Record) Category() *Category {
	return r.category
}

func (r Record) Amount() Money {
	return r.amount
}

func (r Record) DateUTC() time.Time {
	return r.date
}

func (r Record) DateUTCString() time.Time {
	return r.date
}

func (r Record) Type() RecordType {
	return r.recordType
}

func (r Record) BeneficiaryId() AccountId {
	return r.beneficiaryId
}

func (r Record) String() string {
	return fmt.Sprintf("Record{id: %d, type: %s, amount: %s, category: %s, date: %s, beneficiaryId: %d}", r.id, r.recordType, r.amount, r.category, r.DateUTCString(), r.beneficiaryId)
}

type BeneficiaryIdValidator struct {
	Field string
	Value AccountId
	RecordType RecordType
}

func (v *BeneficiaryIdValidator) IsValid(errors *validate.Errors) {
	if v.RecordType == Transfer && v.Value <= 0 {
		errors.Add("beneficiary_id", fmt.Sprintf("beneficiaryId can not be <= 0 when record type is %s", Transfer))
	}
}

type NotNilValidator struct{
	Field string
	Value interface{}
}

func (v *NotNilValidator) IsValid(errors *validate.Errors) {
	if v.Value == nil {
		errors.Add(strings.ToLower(v.Field), fmt.Sprintf("%s is required", v.Field))
	}
}
