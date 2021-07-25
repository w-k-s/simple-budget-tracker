package ledger

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/gobuffalo/validate"
)

type ErrorCode uint64

const (
	ErrUnknown ErrorCode = iota + 1000
	ErrDatabaseConnectivity
	ErrDatabaseState
	ErrUserIdDuplicated
	ErrUserEmailInvalid
	ErrUserEmailDuplicated
	ErrUserNotFound
	ErrAccountValidation
	ErrAccountNotFound
	ErrAccountNameDuplicated
	ErrCurrencyInvalidCode
	ErrCategoryValidation
	ErrCategoryNameDuplicated
	ErrCategoriesNotFound
	ErrRecordValidation
	ErrAmountOverflow
	ErrAmountMismatchingCurrencies
	ErrAmountTotalOfEmptySet
	ErrRequestUnmarshallingFailed
)

var errorCodeNames = map[ErrorCode]string{
	ErrUnknown:                     "UNKOWN",
	ErrDatabaseConnectivity:        "DATABASE_CONNECTIVITY",
	ErrDatabaseState:               "DATABASE_STATE",
	ErrUserIdDuplicated:            "DUPLICATE_USER_ID",
	ErrUserEmailInvalid:            "USER_EMAIL_INVALID",
	ErrUserEmailDuplicated:         "DUPLICATE_USER_EMAIL",
	ErrUserNotFound:                "USER_NOT_FOUND",
	ErrAccountValidation:           "ACCOUNT_VALIDATION_FAILED",
	ErrCurrencyInvalidCode:         "INVALID_CURRENCY_CODE",
	ErrCategoryValidation:          "CATEGORY_VALIDATION_FAILED",
	ErrCategoryNameDuplicated:      "CATEGORY_NAME_DUPLICATED",
	ErrCategoriesNotFound:          "CATEGORIES_NOT_FOUND",
	ErrRecordValidation:            "RECORD_VALIDATION_FAILED",
	ErrAmountOverflow:              "AMOUNT_OVERFLOW",
	ErrAmountMismatchingCurrencies: "AMOUNT_MISMATCHING_CURRENCIES",
	ErrAmountTotalOfEmptySet:       "AMOUNT_TOTAL_OF_EMPTY_SET",
	ErrRequestUnmarshallingFailed:  "REQUEST_UNMARSHALLING_FAILED",
}

func (c ErrorCode) Name() string {
	var name string
	var ok bool
	if name, ok = errorCodeNames[c]; !ok {
		log.Fatalf("No name for error code %d", c)
	}
	return name
}

type Error interface {
	Code() ErrorCode
	Cause() error
	Error() string
	Fields() map[string]string
}

type internalError struct {
	code    ErrorCode
	cause   error
	message string
	fields  map[string]string
}

func (e internalError) Code() ErrorCode {
	return e.code
}

func (e internalError) Cause() error {
	return e.cause
}

func (e internalError) Error() string {
	return e.message
}

func (e internalError) Fields() map[string]string {
	return e.fields
}

func NewError(code ErrorCode, message string, cause error) Error {
	return &internalError{
		code:    code,
		cause:   fmt.Errorf("%w", cause),
		message: message,
		fields:  map[string]string{},
	}
}

func NewErrorWithFields(code ErrorCode, message string, cause error, fields map[string]string) Error {
	return &internalError{
		code:    code,
		cause:   fmt.Errorf("%w", cause),
		message: message,
		fields:  fields,
	}
}

func makeCoreValidationError(code ErrorCode, errors *validate.Errors) error {
	if !errors.HasAny() {
		return nil
	}

	flatErrors := map[string]string{}
	for field, violations := range errors.Errors {
		flatErrors[field] = strings.Join(violations, ", ")
	}

	listErrors := []string{}
	for _, violations := range flatErrors {
		listErrors = append(listErrors, violations)
	}
	sort.Strings(listErrors)

	return NewErrorWithFields(code, strings.Join(listErrors, ", "), nil, flatErrors)
}