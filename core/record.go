package core

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type RecordType string

const (
	Income   RecordType = "INCOME"
	Expense  RecordType = "EXPENSE"
	Transfer RecordType = "TRANSFER"
)

type RecordId uint64
type Record struct {
	id            RecordId
	note          string
	category      *Category
	amount        Money
	date          time.Time
	recordType    RecordType
	beneficiaryId AccountId
}

func NewRecord(id RecordId, note string, category *Category, amount Money, dateUTC time.Time, recordType RecordType, beneficiaryId AccountId) (*Record, error) {
	errors := validate.Validate(
		&validators.IntIsGreaterThan{Name: "Id", Field: int(id), Compared: 0, Message: "Id must be greater than 0"},
		&validators.StringLengthInRange{Name: "Note", Field: note, Min: 0, Max: 50, Message: "Note can not be longer than 50 characters"},
		&categoryValidator{Field: "Category", Value: category},
		&amountValidator{Field: "Amount", Value: amount},
		&validators.TimeIsPresent{Name: "Date", Field: dateUTC, Message: "Invalid date"},
		&validators.FuncValidator{Name: "RecordType", Field: string(recordType), Message: "recordType must be INCOME,EXPENSE or TRANSFER. Invalid: %q", Fn: func() bool {
			for _, rt := range []RecordType{Income, Expense, Transfer} {
				if recordType == rt {
					return true
				}
			}
			return false
		}},
		&beneficiaryIdValidator{Field: "BeneficiaryId", Value: beneficiaryId, RecordType: recordType},
	)

	var err error
	if err = makeCoreValidationError(ErrRecordValidation, errors); err != nil {
		return nil, err
	}

	var actualAmount Money

	if actualAmount, err = amount.Negate(); err != nil {
		return nil, err
	}

	if recordType == Income {
		if actualAmount, err = amount.Abs(); err != nil {
			return nil, err
		}
	}

	record := &Record{
		id:            id,
		note:          note,
		category:      category,
		amount:        actualAmount,
		date:          dateUTC,
		recordType:    recordType,
		beneficiaryId: beneficiaryId,
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

func (r Record) DateUTCString() string {
	return r.date.Format("2006-01-02T15:04:05-0700")
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

type beneficiaryIdValidator struct {
	Field      string
	Value      AccountId
	RecordType RecordType
}

func (v *beneficiaryIdValidator) IsValid(errors *validate.Errors) {
	if v.RecordType == Transfer && v.Value <= 0 {
		errors.Add("beneficiary_id", fmt.Sprintf("beneficiaryId can not be <= 0 when record type is %s", Transfer))
	}
	if v.RecordType != Transfer && v.Value > 0 {
		errors.Add("beneficiary_id", fmt.Sprintf("beneficiaryId must be 0 when record type is %q", v.RecordType))
	}
}

type categoryValidator struct {
	Field string
	Value *Category
}

func (v *categoryValidator) IsValid(errors *validate.Errors) {
	if v.Value == nil {
		errors.Add(strings.ToLower(v.Field), fmt.Sprintf("%s is required", v.Field))
	}
}

type amountValidator struct {
	Field string
	Value Money
}

func (v *amountValidator) IsValid(errors *validate.Errors) {
	if v.Value == nil || (reflect.ValueOf(v.Value).Kind() == reflect.Ptr && reflect.ValueOf(v.Value).IsNil()) {
		errors.Add(strings.ToLower(v.Field), fmt.Sprintf("%s is required", v.Field))
		return
	}
	if v.Value.IsZero() {
		errors.Add(strings.ToLower(v.Field), "amount must not be zero")
	}
}

type Records []*Record

func (rs Records) Total() (Money, error) {
	if rs.Len() == 0 {
		return nil, NewError(ErrAmountTotalOfEmptySet, "No amounts to total", nil)
	}
	total := rs[0].Amount()
	var err error
	for i := 1; i < rs.Len(); i++ {
		total, err = total.Add(rs[i].Amount())
		if err != nil {
			return nil, err
		}
	}
	return total, nil
}

func (rs Records) String() string {
	sort.Sort(rs)
	strs := make([]string, 0, len(rs))
	for _, record := range rs {
		strs = append(strs, record.String())
	}
	return fmt.Sprintf("Records{%s}", strings.Join(strs, ", "))
}

func (rs Records) Len() int           { return len(rs) }
func (rs Records) Swap(i, j int)      { rs[i], rs[j] = rs[j], rs[i] }
func (rs Records) Less(i, j int) bool { return rs[i].DateUTC().Unix() < rs[j].DateUTC().Unix() }
