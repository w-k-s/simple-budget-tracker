package pkg

import (
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/gobuffalo/validate"
)

type ValidationError struct {
	code   ErrorCode
	title  string
	detail string
	cause  error
	fields map[string]string
}

func (i ValidationError) Unwrap() error {
	return i.cause
}

func (i ValidationError) Error() string {

	sb := strings.Builder{}

	if len(i.detail) > 0 {
		sb.WriteString(i.detail)
		if !strings.HasSuffix(i.detail, ".") {
			sb.WriteString(".")
		}
	}

	if i.cause != nil {
		cause := i.cause.Error()
		sb.WriteString(" Reason: ")
		sb.WriteString(cause)

		if !strings.HasSuffix(cause, ".") {
			sb.WriteString(".")
		}
	}

	if len(i.fields) > 0 {
		fieldErrors := []string{}
		for _, fieldError := range i.fields {
			fieldErrors = append(fieldErrors, fieldError)
		}
		sort.Strings(fieldErrors)

		sb.WriteString(strings.Join(fieldErrors, ","))
	}

	return sb.String()
}

func (i ValidationError) Code() uint64 {
	return uint64(i.code)
}

func (i ValidationError) Title() string {
	return i.title
}

func (i ValidationError) Detail() string {
	return i.detail
}

func (i ValidationError) InvalidFields() map[string]string {
	return i.fields
}

func (i ValidationError) StatusCode() int {
	return i.code.status()
}

func ValidationErrorWithErrors(
	code ErrorCode,
	message string,
	errors *validate.Errors,
) error {
	if !errors.HasAny() {
		return nil
	}

	flatErrors := map[string]string{}
	for field, violations := range errors.Errors {
		flatErrors[field] = strings.Join(violations, ", ")
	}

	return ValidationError{
		code:   code,
		title:  code.name(),
		detail: message,
		cause:  nil,
		fields: flatErrors,
	}
}

func ValidationErrorWithFields(
	code ErrorCode,
	message string,
	err error,
	errors map[string]string,
) error {

	return ValidationError{
		code:   code,
		title:  code.name(),
		detail: message,
		cause:  nil,
		fields: errors,
	}
}

func ValidationErrorWithError(
	code ErrorCode,
	message string,
	err error,
) error {

	return ValidationError{
		code:   code,
		title:  code.name(),
		detail: message,
		cause:  err,
		fields: nil,
	}
}

type SystemError struct {
	code   ErrorCode
	title  string
	detail string
	cause  error
}

func NewSystemError(
	code ErrorCode,
	message string,
	cause error,
) error {
	return SystemError{
		code:   code,
		title:  code.name(),
		detail: message,
		cause:  cause,
	}
}

func (s SystemError) Unwrap() error {
	return s.cause
}

func (s SystemError) Error() string {
	sb := strings.Builder{}

	if len(s.detail) > 0 {
		sb.WriteString(s.detail)
		if !strings.HasSuffix(s.detail, ".") {
			sb.WriteString(".")
		}
	}

	if s.cause != nil {
		sb.WriteString(" Reason: ")
		sb.WriteString(s.cause.Error())
	}

	return sb.String()
}

func (s SystemError) Code() uint64 {
	return uint64(s.code)
}

func (s SystemError) Title() string {
	return s.title
}

func (s SystemError) Detail() string {
	return s.detail
}

func (s SystemError) StatusCode() int {
	return s.code.status()
}

// --

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
	ErrBudgetValidation
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
	ErrBudgetValidation:            "BUDGET_VALIDATION_FAILED",
}

func (c ErrorCode) name() string {
	var name string
	var ok bool
	if name, ok = errorCodeNames[c]; !ok {
		log.Fatalf("FATAL: No name for error code %d", c)
	}
	return name
}

func (c ErrorCode) status() int {
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
