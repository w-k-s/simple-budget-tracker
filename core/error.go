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
)

var errorCodeNames = map[ErrorCode]string{
	ErrUnknown:              "UNKOWN",
	ErrDatabaseConnectivity: "DATABASE_CONNECTIVITY",
	ErrDatabaseState:        "DATABASE_STATE",
	ErrDuplicateUserId:      "DUPLICATE_USER_ID",
	ErrDuplicateUserEmail:   "DUPLICATE_USER_EMAIL",
	ErrUserNotFound:         "USER_NOT_FOUND",
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
	Fields() map[string]interface{}
}

type internalError struct {
	code    ErrorCode
	cause   error
	message string
	fields  map[string]interface{}
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

func (e internalError) Fields() map[string]interface{} {
	return e.fields
}

func NewError(code ErrorCode, message string, cause error) Error {
	return &internalError{
		code:    code,
		cause:   fmt.Errorf("%w", cause),
		message: message,
		fields:  map[string]interface{}{},
	}
}
