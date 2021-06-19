package core

import "github.com/bojanz/currency"

func IsValidCurrency(currencyCode string) bool {
	return currency.IsValid(currencyCode)
}
