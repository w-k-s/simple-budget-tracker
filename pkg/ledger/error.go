package ledger

import (
	"fmt"
	"log"
	"net/http"
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
	ErrRecordsPeriodOfEmptySet
	ErrAmountOverflow
	ErrAmountMismatchingCurrencies
	ErrAmountTotalOfEmptySet
	ErrAuditValidation
	ErrAuditUpdatedByBadFormat
	ErrRequestUnmarshallingFailed
	ErrServiceUserIdRequired
	ErrServiceAccountIdRequired
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
	ErrAccountNotFound:             "ACCOUNT_NOT_FOUND",
	ErrAccountNameDuplicated:       "ACCOUNT_NAME_DUPLICATED",
	ErrCurrencyInvalidCode:         "INVALID_CURRENCY_CODE",
	ErrCategoryValidation:          "CATEGORY_VALIDATION_FAILED",
	ErrCategoryNameDuplicated:      "CATEGORY_NAME_DUPLICATED",
	ErrCategoriesNotFound:          "CATEGORIES_NOT_FOUND",
	ErrRecordValidation:            "RECORD_VALIDATION_FAILED",
	ErrRecordsPeriodOfEmptySet:     "RECORDS_PERIOD_OF_EMPTY_SET",
	ErrAmountOverflow:              "AMOUNT_OVERFLOW",
	ErrAmountMismatchingCurrencies: "AMOUNT_MISMATCHING_CURRENCIES",
	ErrAmountTotalOfEmptySet:       "AMOUNT_TOTAL_OF_EMPTY_SET",
	ErrAuditValidation:             "AUDIT_VALIDATION_FAILED",
	ErrAuditUpdatedByBadFormat:     "AUDIT_UPDATED_BY_BAD_FORMAT",
	ErrRequestUnmarshallingFailed:  "REQUEST_UNMARSHALLING_FAILED",
	ErrServiceUserIdRequired:       "SERVICE_REQUIRED_USER_ID",
	ErrServiceAccountIdRequired:    "SERVICE_REQUIRED_ACCOUNT_ID",
}

func (c ErrorCode) Name() string {
	var name string
	var ok bool
	if name, ok = errorCodeNames[c]; !ok {
		log.Fatalf("FATAL: No name for error code %d", c)
	}
	return name
}

func (c ErrorCode) Status() int {
	switch c {
	case ErrUserIdDuplicated:
		fallthrough
	case ErrUserEmailInvalid:
		fallthrough
	case ErrUserEmailDuplicated:
		fallthrough
	case ErrAccountValidation:
		fallthrough
	case ErrAccountNameDuplicated:
		fallthrough
	case ErrCurrencyInvalidCode:
		fallthrough
	case ErrCategoryValidation:
		fallthrough
	case ErrCategoryNameDuplicated:
		fallthrough
	case ErrRecordValidation:
		fallthrough
	case ErrAmountMismatchingCurrencies:
		fallthrough
	case ErrAuditValidation:
		fallthrough
	case ErrAuditUpdatedByBadFormat:
		fallthrough
	case ErrRequestUnmarshallingFailed:
		fallthrough
	case ErrServiceAccountIdRequired:
		return http.StatusBadRequest

	case ErrServiceUserIdRequired:
		return http.StatusUnauthorized

	case ErrUserNotFound:
		fallthrough
	case ErrAccountNotFound:
		fallthrough
	case ErrCategoriesNotFound:
		return http.StatusNotFound

	case ErrDatabaseConnectivity:
		fallthrough
	case ErrDatabaseState:
		fallthrough
	case ErrAmountTotalOfEmptySet:
		fallthrough
	case ErrRecordsPeriodOfEmptySet:
		fallthrough
	case ErrAmountOverflow:
		fallthrough
	case ErrUnknown:
		fallthrough
	default:
		return http.StatusInternalServerError
	}
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
