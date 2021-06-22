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
