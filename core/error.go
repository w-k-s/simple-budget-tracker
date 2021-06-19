package core

import (
	"fmt"
	"log"
)

type ErrorCode uint64

const (
	ErrUnknown ErrorCode = iota + 1000
	ErrDatabaseConnectivity
	ErrDatabaseState
	ErrDuplicateUserId
	ErrDuplicateUserEmail
	ErrUserNotFound
	ErrAccountValidation
)

var errorCodeNames = map[ErrorCode]string{
	ErrUnknown:              "UNKOWN",
	ErrDatabaseConnectivity: "DATABASE_CONNECTIVITY",
	ErrDatabaseState:        "DATABASE_STATE",
	ErrDuplicateUserId:      "DUPLICATE_USER_ID",
	ErrDuplicateUserEmail:   "DUPLICATE_USER_EMAIL",
	ErrUserNotFound:         "USER_NOT_FOUND",
	ErrAccountValidation:    "ACCOUNT_VALIDATION_FAILED",
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
