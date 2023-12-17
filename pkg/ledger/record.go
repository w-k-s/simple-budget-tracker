package ledger

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/google/uuid"
)

type RecordType string

const (
	// Recording amount credited to account
	Income RecordType = "INCOME"
	// Recording amount debited from account
	Expense RecordType = "EXPENSE"
	// Recording amount transferred between accounts (belonging to the same user)
	Transfer RecordType = "TRANSFER"
)

const (
	NoSourceAccount      = AccountId(0)
	NoBeneficiaryAccount = AccountId(0)
	NoBeneficiaryType    = AccountType("")
	NoTransferReference  = TransferReference("")
)

func NoTransfer() (AccountId, AccountId, TransferReference) {
	return NoSourceAccount, NoBeneficiaryAccount, NoTransferReference
}

// TransferReference is a value that's the same for both the credit and debit records of a transfer
type TransferReference string

func MakeTransferReference() TransferReference {
	return TransferReference(uuid.NewString())
}

type RecordId uint64
type Record struct {
	auditInfo
	id                RecordId
	note              string
	category          Category
	amount            Money
	date              time.Time
	recordType        RecordType
	sourceAccountId   AccountId
	beneficiaryId     AccountId
	beneficiaryType   AccountType
	transferReference TransferReference
}

// I did not think the naming through :(
type RecordRecord interface {
	Id() RecordId
	Note() string
	Category() Category
	Amount() Money
	DateUTC() time.Time
	RecordType() RecordType
	SourceAccountId() AccountId
	BeneficiaryId() AccountId
	BeneficiaryType() AccountType
	TransferReference() TransferReference
	CreatedBy() UpdatedBy
	CreatedAtUTC() time.Time
	ModifiedBy() UpdatedBy
	ModifiedAtUTC() time.Time
	Version() Version
}

func NewRecord(
	id RecordId,
	note string,
	category Category,
	amount Money,
	dateUTC time.Time,
	recordType RecordType,
	sourceAccountId AccountId,
	beneficiaryId AccountId,
	beneficiaryType AccountType,
	transferReference TransferReference,
	updatedBy UpdatedBy,
) (Record, error) {
	var (
		auditInfo auditInfo
		err       error
	)

	if auditInfo, err = makeAuditForCreation(updatedBy); err != nil {
		return Record{}, err
	}

	return newRecord(
		id,
		note,
		category,
		amount,
		dateUTC,
		recordType,
		sourceAccountId,
		beneficiaryId,
		beneficiaryType,
		transferReference,
		auditInfo,
	)
}

func NewRecordFromRecord(rr RecordRecord) (Record, error) {
	var (
		auditInfo auditInfo
		err       error
	)

	if auditInfo, err = makeAuditForModification(
		rr.CreatedBy(),
		rr.CreatedAtUTC(),
		rr.ModifiedBy(),
		rr.ModifiedAtUTC(),
		rr.Version(),
	); err != nil {
		return Record{}, err
	}

	return newRecord(
		rr.Id(),
		rr.Note(),
		rr.Category(),
		rr.Amount(),
		rr.DateUTC(),
		rr.RecordType(),
		rr.SourceAccountId(),
		rr.BeneficiaryId(),
		rr.BeneficiaryType(),
		rr.TransferReference(),
		auditInfo,
	)
}

func newRecord(
	id RecordId,
	note string,
	category Category,
	amount Money,
	dateUTC time.Time,
	recordType RecordType,
	sourceAccountId,
	beneficiaryId AccountId,
	beneficiaryType AccountType,
	transferReference TransferReference,
	auditInfo auditInfo,
) (Record, error) {
	errors := validate.Validate(
		&validators.IntIsGreaterThan{Name: "Id", Field: int(id), Compared: 0, Message: "Id must be greater than 0"},
		&validators.StringLengthInRange{Name: "Note", Field: note, Min: 0, Max: 50, Message: "Note can not be longer than 50 characters"},
		&categoryValidator{Field: "Category", Value: category},
		&amountValidator{Field: "Amount", Value: amount},
		&validators.TimeIsPresent{Name: "Date", Field: dateUTC, Message: "Invalid date"},
		&validators.StringInclusion{Name: "RecordType", Field: string(recordType), List: []string{"INCOME", "EXPENSE", "TRANSFER"}, Message: "recordType must be INCOME,EXPENSE or TRANSFER."},
		&beneficiaryIdValidator{BeneficiaryId: beneficiaryId, SourceAccountId: sourceAccountId, RecordType: recordType},
		&beneficiaryTypeValidator{Field: string(beneficiaryType)},
		&transferReferenceValidator{Value: transferReference, RecordType: recordType},
	)

	var err error
	if err = makeCoreValidationError(ErrRecordValidation, errors); err != nil {
		return Record{}, err
	}

	actualAmount := amount

	if recordType == Expense {
		if actualAmount, err = amount.Negate(); err != nil {
			return Record{}, err
		}
	}

	if recordType == Income {
		if actualAmount, err = amount.Abs(); err != nil {
			return Record{}, err
		}
	}

	record := Record{
		auditInfo:         auditInfo,
		id:                id,
		note:              note,
		category:          category,
		amount:            actualAmount,
		date:              dateUTC,
		recordType:        recordType,
		sourceAccountId:   sourceAccountId,
		beneficiaryId:     beneficiaryId,
		beneficiaryType:   beneficiaryType,
		transferReference: transferReference,
	}

	return record, nil
}

func (r Record) Id() RecordId {
	return r.id
}

func (r Record) Note() string {
	return r.note
}

func (r Record) Category() Category {
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

func (r Record) SourceAccountId() AccountId {
	return r.sourceAccountId
}

func (r Record) BeneficiaryId() AccountId {
	return r.beneficiaryId
}

func (r Record) BeneficiaryType() AccountType {
	return r.beneficiaryType
}

func (r Record) TransferReference() TransferReference {
	return r.transferReference
}

func (r Record) IsTransferToSavingAccount() bool {
	return r.recordType == Transfer && r.beneficiaryType == AccountTypeSaving
}

func (r Record) String() string {
	return fmt.Sprintf("Record{id: %d, type: %s, amount: %s, category: %s, date: %s, sourceAccountId: %d, beneficiaryId: %d, beneficiaryType: %s, transferReference: %s}",
		r.id,
		r.recordType,
		r.amount,
		r.category,
		r.DateUTCString(),
		r.sourceAccountId,
		r.beneficiaryId,
		r.beneficiaryType,
		r.transferReference,
	)
}

type beneficiaryIdValidator struct {
	BeneficiaryId   AccountId
	SourceAccountId AccountId
	RecordType      RecordType
}

func (v *beneficiaryIdValidator) IsValid(errors *validate.Errors) {
	if v.RecordType == Transfer && v.BeneficiaryId <= 0 {
		errors.Add("beneficiaryId", fmt.Sprintf("beneficiaryId can not be <= 0 when record type is %s", Transfer))
	}
	if v.RecordType == Transfer && v.SourceAccountId <= 0 {
		errors.Add("sourceAccountId", fmt.Sprintf("sourceAccountId can not be <= 0 when record type is %s", Transfer))
	}
	if v.RecordType != Transfer && v.BeneficiaryId > 0 {
		errors.Add("beneficiaryId", fmt.Sprintf("beneficiaryId must be 0 when record type is %q", v.RecordType))
	}
	if v.RecordType != Transfer && v.SourceAccountId > 0 {
		errors.Add("sourceAccountId", fmt.Sprintf("sourceAccountId must be 0 when record type is %q", v.RecordType))
	}
}

type beneficiaryTypeValidator struct {
	Field      string
	RecordType RecordType
}

func (v *beneficiaryTypeValidator) IsValid(errors *validate.Errors) {
	if v.RecordType != Transfer {
		return
	}
	validator := &accountTypeValidator{
		Name:  "BeneficiaryType",
		Field: v.Field,
	}
	validator.IsValid(errors)
}

type transferReferenceValidator struct {
	Value      TransferReference
	RecordType RecordType
}

func (v *transferReferenceValidator) IsValid(errors *validate.Errors) {
	if v.RecordType == Transfer && len(v.Value) == 0 {
		errors.Add("transferReference", fmt.Sprintf("transferReference can not be empty when record type is %s", Transfer))
	}
	if v.RecordType != Transfer && len(v.Value) > 0 {
		errors.Add("transferReference", fmt.Sprintf("transferReference must be empty when record type is %s", v.RecordType))
	}
}

type categoryValidator struct {
	Field string
	Value Category
}

func (v *categoryValidator) IsValid(errors *validate.Errors) {
	var c Category
	if v.Value == c {
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

type Records []Record

// ====== RECORD CALCULATIONS =============
//  Each time a records calculation method is called, the entire slice is looped over.
//  TODO: Calculate everything at once and cache.
//  TODO++: Perform these calculations in a dao? e.g. PeriodDao
// ========================================

// Net Balance of given records = Total Income - Total Expenses
func (rs Records) NetBalance() (Money, error) {
	if rs.Len() == 0 {
		return nil, NewError(ErrAmountTotalOfEmptySet, "No amounts to total", nil)
	}

	var (
		total, _ = NewMoney(rs[0].Amount().Currency().CurrencyCode(), 0)
		err      error
	)
	for i := 0; i < rs.Len(); i++ {
		total, err = total.Add(rs[i].Amount())
		if err != nil {
			return nil, err
		}
	}
	return total, nil
}

// Total Expenses of given records = Total Expenses + Transfers of given records
func (rs Records) TotalExpenses() (Money, error) {
	if rs.Len() == 0 {
		return nil, NewError(ErrAmountTotalOfEmptySet, "No amounts to total", nil)
	}

	var (
		total, _  = NewMoney(rs[0].Amount().Currency().CurrencyCode(), 0)
		amountAbs Money
		err       error
	)
	for i := 0; i < rs.Len(); i++ {
		record := rs[i]
		if record.recordType == Expense {
			amountAbs, err = record.Amount().Abs()
			if err != nil {
				return nil, err
			}
			total, err = total.Add(amountAbs)
			if err != nil {
				return nil, err
			}
		}
	}
	return total, nil
}

// Total income of given records
func (rs Records) TotalIncome() (Money, error) {
	if rs.Len() == 0 {
		return nil, NewError(ErrAmountTotalOfEmptySet, "No amounts to total", nil)
	}

	total, _ := NewMoney(rs[0].Amount().Currency().CurrencyCode(), 0)
	var err error
	for i := 0; i < rs.Len(); i++ {
		record := rs[i]
		if record.recordType == Income {
			total, err = total.Add(record.Amount())
			if err != nil {
				return nil, err
			}
		}
	}
	return total, nil
}

// Total saved
func (rs Records) TotalSavings() (Money, error) {
	if rs.Len() == 0 {
		return nil, NewError(ErrAmountTotalOfEmptySet, "No amounts to total", nil)
	}

	total, _ := NewMoney(rs[0].Amount().Currency().CurrencyCode(), 0)
	var (
		amountAbs Money
		err       error
	)

	for i := 0; i < rs.Len(); i++ {
		record := rs[i]

		if record.IsTransferToSavingAccount() {
			amountAbs, err = record.Amount().Abs()

			if err != nil {
				return nil, err
			}
			total, err = total.Add(amountAbs)
			if err != nil {
				return nil, err
			}
		}
	}
	return total, nil
}

// Expenses - Savings
func (rs Records) NetExpenses() (Money, error) {
	// TODO
	if rs.Len() == 0 {
		return nil, NewError(ErrAmountTotalOfEmptySet, "No amounts to total", nil)
	}

	total, _ := NewMoney(rs[0].Amount().Currency().CurrencyCode(), 0)
	var (
		amountAbs Money
		err       error
	)
	for i := 0; i < rs.Len(); i++ {
		record := rs[i]
		if record.recordType == Expense {
			amountAbs, err = record.Amount().Abs()
			if err != nil {
				return nil, err
			}
			total, err = total.Add(amountAbs)
			if err != nil {
				return nil, err
			}
		}
	}
	return total, nil
}

func (rs Records) Period() (time.Time, time.Time, error) {
	if len(rs) == 0 {
		return time.Time{}, time.Time{}, NewError(ErrRecordsPeriodOfEmptySet, "Can not determine records period for empty set", nil)
	}

	var (
		from = rs[0].date
		to   = rs[0].date
	)

	for _, r := range rs {
		if r.date.After(to) {
			to = r.date
		}
		if r.date.Before(from) {
			from = r.date
		}
	}

	return from, to, nil
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
