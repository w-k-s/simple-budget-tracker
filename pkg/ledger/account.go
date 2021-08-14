package ledger

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type AccountId uint64
type Account struct {
	auditInfo
	id       AccountId
	name     string
	currency string
}

type AccountRecord interface {
	Id() AccountId
	Name() string
	Currency() string
	CreatedBy() UpdatedBy
	CreatedAtUTC() time.Time
	ModifiedBy() UpdatedBy
	ModifiedAtUTC() time.Time
	Version() Version
}

func NewAccount(id AccountId, name string, currency string, createdBy UpdatedBy) (Account, error) {
	var (
		auditInfo auditInfo
		err       error
	)

	if auditInfo, err = makeAuditForCreation(createdBy); err != nil {
		return Account{}, err
	}

	return newAccount(id, name, currency, auditInfo)
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

	return newAccount(record.Id(), record.Name(), record.Currency(), auditInfo)
}

func newAccount(id AccountId, name string, currency string, auditInfo auditInfo) (Account, error) {

	errors := validate.Validate(
		&validators.IntIsGreaterThan{Name: "Id", Field: int(id), Compared: 0, Message: "Id must be greater than 0"},
		&validators.StringLengthInRange{Name: "Name", Field: name, Min: 1, Max: 25, Message: "Name must be 1 and 25 characters long"},
		&validators.StringLengthInRange{Name: "Currency", Field: currency, Min: 3, Max: 4, Message: "Currency must be 3 characters long"},
		&validators.FuncValidator{Name: "Currency", Field: currency, Message: "No such currency '%s'", Fn: func() bool { return IsValidCurrency(currency) }},
	)

	if err := makeCoreValidationError(ErrAccountValidation, errors); err != nil {
		return Account{}, err
	}

	return Account{
		auditInfo: auditInfo,
		id:        id,
		name:      name,
		currency:  currency,
	}, nil
}

func (a Account) Id() AccountId {
	return a.id
}

func (a Account) Name() string {
	return a.name
}

func (a Account) Currency() string {
	return a.currency
}

func (a Account) String() string {
	return fmt.Sprintf("Account{id: %d, name: %s, currency: %s}", a.id, a.name, a.currency)
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
	for _, category := range as {
		strs = append(strs, category.String())
	}
	return fmt.Sprintf("Accounts{%s}", strings.Join(strs, ", "))
}

func (as Accounts) Len() int           { return len(as) }
func (as Accounts) Swap(i, j int)      { as[i], as[j] = as[j], as[i] }
func (as Accounts) Less(i, j int) bool { return as[i].Name() < as[j].Name() }
