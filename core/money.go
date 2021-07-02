package core

import (
	"fmt"
	"strconv"

	"github.com/bojanz/currency"
)

func IsValidCurrency(currencyCode string) bool {
	if len(currencyCode) != 3 {
		return false
	}
	return currency.IsValid(currencyCode)
}

type Currency interface {
	Code() string
}

type Money interface {
	Currency() Currency
	IsPositive() bool
	IsZero() bool
	IsNegative() bool
	Abs() (Money, error)
	Negate() (Money, error)
	MinorUnits() (int64, error)
	String() string
}

type internalCurrency struct {
	code string
}

func (i internalCurrency) Code() string {
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

func (i internalMoney) Abs() (Money, error) {
	if i.IsNegative() {
		minorUnits, err := i.MinorUnits()
		if err != nil {
			return nil, NewError(ErrAmountOverflow, "The number is too large to be represented", err)
		}
		return NewMoney(i.Currency().Code(), -1*minorUnits)
	}
	return i, nil
}

func (i internalMoney) Negate() (Money, error) {
	if i.IsPositive() {
		minorUnits, err := i.MinorUnits()
		if err != nil {
			return nil, NewError(ErrAmountOverflow, "The number is too large to be represented", err)
		}
		return NewMoney(i.Currency().Code(), -1*minorUnits)
	}
	return i, nil
}

func (i internalMoney) MinorUnits() (int64, error) {
	return i.amount.Int64()
}

func (i internalMoney) String() string {
	return fmt.Sprintf("%s %s", i.amount.CurrencyCode(), i.amount.Number())
}

func NewMoney(currencyCode string, amountMinorUnits int64) (Money, error) {
	amount, err := currency.NewAmountFromInt64(amountMinorUnits, currencyCode)
	if err != nil {
		if _, ok := err.(currency.InvalidCurrencyCodeError); ok {
			return nil, NewErrorWithFields(ErrCurrencyInvalidCode, err.Error(), err, map[string]string{"code": currencyCode})
		}
		return nil, NewErrorWithFields(ErrUnknown, "Invalid Monetary amount", err, map[string]string{"code": currencyCode, "amount": strconv.FormatInt(amountMinorUnits, 10)})
	}
	return &internalMoney{amount}, nil
}
