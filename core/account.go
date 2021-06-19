package core

import (
	"fmt"
	"strings"

	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type AccountId uint64
type Account struct {
	id       AccountId
	name     string
	currency string
}

func NewAccount(id AccountId, name string, currency string) (*Account, error) {
	account := &Account{
		id:       id,
		name:     name,
		currency: currency,
	}

	errors := validate.Validate(
		&validators.IntIsGreaterThan{Name: "Id", Field: int(account.id), Compared: 0, Message: "Id must be greater than 0"},
		&validators.StringLengthInRange{Name: "Name", Field: account.name, Min: 1, Max: 255, Message: "Name must be 1 and 255 characters long"},
		&validators.StringLengthInRange{Name: "Currency", Field: account.currency, Min: 3, Max: 4, Message: "Currency must be 3 characters long"},
		&validators.FuncValidator{Name: "Currency", Field: account.currency, Message: "No such currency %q", Fn: func() bool { return IsValidCurrency(account.currency) }},
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
		return nil, NewErrorWithFields(ErrAccountValidation, strings.Join(listErrors, ", "), errors, flatErrors)
	}
	return account, nil
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
