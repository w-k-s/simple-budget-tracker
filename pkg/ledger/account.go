package ledger

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/w-k-s/simple-budget-tracker/pkg"
)

type AccountId uint64
type AccountType string

const (
	AccountTypeCurrent AccountType = "Current"
	AccountTypeSaving  AccountType = "Saving"
)

type Account struct {
	auditInfo
	id             AccountId
	name           string
	currency       string
	currentBalance Money
	accountType    AccountType
}

type AccountRecord interface {
	Id() AccountId
	Name() string
	Type() AccountType
	Currency() string
	CurrentBalanceMinorUnits() int64
	CreatedBy() UpdatedBy
	CreatedAtUTC() time.Time
	ModifiedBy() UpdatedBy
	ModifiedAtUTC() time.Time
	Version() Version
}

func NewAccount(
	id AccountId,
	name string,
	accountType AccountType,
	currency string,
	createdBy UpdatedBy,
) (Account, error) {
	var (
		auditInfo auditInfo
		err       error
	)

	if auditInfo, err = makeAuditForCreation(createdBy); err != nil {
		return Account{}, err
	}

	return newAccount(id, name, accountType, currency, 0, auditInfo)
}

func NewAccountFromRecord(record AccountRecord) (Account, error) {
	var (
		auditInfo auditInfo
		err       error
	)

	if auditInfo, err = makeAuditForModification(
		record.CreatedBy(),
		record.CreatedAtUTC(),
		record.ModifiedBy(),
		record.ModifiedAtUTC(),
		record.Version(),
	); err != nil {
		return Account{}, err
	}

	return newAccount(record.Id(), record.Name(), record.Type(), record.Currency(), record.CurrentBalanceMinorUnits(), auditInfo)
}

func newAccount(
	id AccountId,
	name string,
	accountType AccountType,
	currency string,
	currentBalanceMinorUnits int64,
	auditInfo auditInfo,
) (Account, error) {

	errors := validate.Validate(
		&validators.IntIsGreaterThan{
			Name:     "Id",
			Field:    int(id),
			Compared: 0,
			Message:  "Id must be greater than 0",
		},
		&validators.StringLengthInRange{
			Name:    "Name",
			Field:   name,
			Min:     1,
			Max:     25,
			Message: "Name must be 1 and 25 characters long",
		},
		&accountTypeValidator{
			Name:  "AccountType",
			Field: string(accountType),
		},
		&currencyValidator{
			Name:  "Currency",
			Value: currency,
		},
	)

	err := pkg.ValidationErrorWithErrors(pkg.ErrAccountValidation, "", errors)
	if err != nil {
		return Account{}, err
	}

	currentBalance, err := NewMoney(currency, currentBalanceMinorUnits)
	if err != nil {
		return Account{}, err
	}

	return Account{
		auditInfo:      auditInfo,
		id:             id,
		name:           strings.Title(strings.ToLower(name)),
		accountType:    AccountType(accountType),
		currency:       currency,
		currentBalance: currentBalance,
	}, nil
}

func (a Account) Id() AccountId {
	return a.id
}

func (a Account) Name() string {
	return a.name
}

func (a Account) Type() AccountType {
	return a.accountType
}

func (a Account) Currency() string {
	return a.currency
}

func (a Account) CurrentBalance() Money {
	return a.currentBalance
}

func (a Account) String() string {
	return fmt.Sprintf("Account{id: %d, name: %s, type: %s, currency: %s, balance: %s}",
		a.id,
		a.name,
		a.accountType,
		a.currency,
		a.currentBalance.String(),
	)
}

type Accounts []Account

func (as Accounts) Names() []string {
	names := make([]string, 0, len(as))
	for _, account := range as {
		names = append(names, account.Name())
	}
	sort.Strings(names)
	return names
}

func (as Accounts) String() string {
	sort.Sort(as)
	strs := make([]string, 0, len(as))
	for _, account := range as {
		strs = append(strs, account.String())
	}
	return fmt.Sprintf("Accounts{%s}", strings.Join(strs, ", "))
}

func (as Accounts) Len() int           { return len(as) }
func (as Accounts) Swap(i, j int)      { as[i], as[j] = as[j], as[i] }
func (as Accounts) Less(i, j int) bool { return as[i].Name() < as[j].Name() }

type currencyValidator struct {
	Name  string
	Value string
}

func (v *currencyValidator) IsValid(errors *validate.Errors) {
	if len(v.Value) == 0 {
		errors.Add(strings.ToLower(v.Name), "currency is required")
		return
	}
	if !IsValidCurrency(v.Value) {
		errors.Add(strings.ToLower(v.Name), fmt.Sprintf("No such currency '%s'", v.Value))
	}
}

type accountTypeValidator struct {
	Name  string
	Field string
}

func (v *accountTypeValidator) IsValid(errors *validate.Errors) {
	if len(v.Field) == 0 {
		errors.Add(strings.ToLower(v.Name), "accountType is required")
		return
	}
	validAccountTypes := []string{
		string(AccountTypeCurrent),
		string(AccountTypeSaving),
	}

	validator := &validators.StringInclusion{
		Name:  v.Name,
		Field: v.Field,
		List: []string{
			string(AccountTypeCurrent),
			string(AccountTypeSaving),
		},
		Message: fmt.Sprintf("account type must be one of %q", validAccountTypes),
	}
	validator.IsValid(errors)
}
