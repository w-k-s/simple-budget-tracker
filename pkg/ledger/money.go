package ledger

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/bojanz/currency"
	"github.com/gobuffalo/validate"
	"github.com/w-k-s/simple-budget-tracker/pkg"
)

func IsValidCurrency(currencyCode string) bool {
	if len(currencyCode) != 3 {
		return false
	}
	return currency.IsValid(currencyCode)
}

type Currency interface {
	CurrencyCode() string
}

type Money interface {
	Currency() Currency
	IsPositive() bool
	IsZero() bool
	IsNegative() bool
	Add(m Money) (Money, error)
	Abs() (Money, error)
	Negate() (Money, error)

	MinorUnits() (int64, error)
	MustMinorUnits() int64

	String() string
}

type internalCurrency struct {
	code string
}

func (i internalCurrency) CurrencyCode() string {
	return i.code
}

type internalMoney struct {
	amount currency.Amount
}

func (i internalMoney) Currency() Currency {
	return internalCurrency{i.amount.CurrencyCode()}
}

func (i internalMoney) IsZero() bool {
	return i.amount.IsZero()
}

func (i internalMoney) IsPositive() bool {
	return i.amount.IsPositive()
}

func (i internalMoney) IsNegative() bool {
	return i.amount.IsNegative()
}

func (i internalMoney) Add(m Money) (Money, error) {
	if i.Currency().CurrencyCode() != m.Currency().CurrencyCode() {
		return nil, pkg.ValidationErrorWithFields(pkg.ErrAmountMismatchingCurrencies, "Can not sum mismatching currencies", nil, nil)
	}
	var (
		leftMinorUnits  int64
		rightMinorUnits int64
		err             error
	)

	if leftMinorUnits, err = i.MinorUnits(); err != nil {
		return nil, err
	}
	if rightMinorUnits, err = m.MinorUnits(); err != nil {
		return nil, err
	}
	return NewMoney(i.Currency().CurrencyCode(), leftMinorUnits+rightMinorUnits)
}

func (i internalMoney) Abs() (Money, error) {
	if i.IsNegative() {
		var minorUnits int64
		var err error
		if minorUnits, err = i.MinorUnits(); err != nil {
			return nil, pkg.ValidationErrorWithError(pkg.ErrAmountOverflow, "The number is too large to be represented", err)
		}
		return NewMoney(i.Currency().CurrencyCode(), -1*minorUnits)
	}
	return i, nil
}

func (i internalMoney) Negate() (Money, error) {
	if i.IsPositive() {
		var minorUnits int64
		var err error
		if minorUnits, err = i.MinorUnits(); err != nil {
			return nil, pkg.ValidationErrorWithError(pkg.ErrAmountOverflow, "The number is too large to be represented", err)
		}
		return NewMoney(i.Currency().CurrencyCode(), -1*minorUnits)
	}
	return i, nil
}

func (i internalMoney) MinorUnits() (int64, error) {
	return i.amount.Int64()
}

func (i internalMoney) MustMinorUnits() int64 {
	minorUnits, err := i.amount.Int64()
	if err != nil {
		log.Fatalf("Failed to convert '%v' to minor units", i.amount)
	}
	return minorUnits
}

func (i internalMoney) String() string {
	return fmt.Sprintf("%s %s", i.amount.CurrencyCode(), i.amount.Number())
}

func NewMoney(currencyCode string, amountMinorUnits int64) (Money, error) {
	amount, err := currency.NewAmountFromInt64(amountMinorUnits, currencyCode)
	if err != nil {
		if _, ok := err.(currency.InvalidCurrencyCodeError); ok {
			return nil, pkg.ValidationErrorWithFields(pkg.ErrCurrencyInvalidCode, err.Error(), err, map[string]string{"code": currencyCode})
		}
		return nil, pkg.ValidationErrorWithFields(pkg.ErrUnknown, "Invalid Monetary amount", err, map[string]string{"code": currencyCode, "amount": strconv.FormatInt(amountMinorUnits, 10)})
	}
	return &internalMoney{amount}, nil
}

func MustMoney(m Money, err error) Money {
	if err != nil {
		log.Fatal(err)
	}
	return m
}

type amountPositiveOrZeroValidator struct {
	Name    string
	Field   Money
	Message string
}

func (v *amountPositiveOrZeroValidator) IsValid(errors *validate.Errors) {
	amount := v.Field.MustMinorUnits()
	if amount < 0 {
		errors.Add(strings.ToLower(v.Name), v.Message)
	}
}
