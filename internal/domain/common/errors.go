package common

import (
	"errors"
	"fmt"
)

type ErrorCode string

const (
	CodeInvalid          ErrorCode = "invalid_argument"
	CodeNotFound         ErrorCode = "not_found"
	CodeConflict         ErrorCode = "conflict"
	CodeForbidden        ErrorCode = "forbidden"
	CodeUnauthenticated  ErrorCode = "unauthenticated"
	CodeExpired          ErrorCode = "expired"
	CodeAlreadyProcessed ErrorCode = "already_processed"
	CodeCapacityFull     ErrorCode = "capacity_full"
	CodeVersionConflict  ErrorCode = "version_conflict"
	CodeRateLimited      ErrorCode = "rate_limited"
	CodeInternal         ErrorCode = "internal"
)

type DomainError struct {
	Code    ErrorCode
	Message string
	Fields  map[string]string
	Cause   error
}

func (e *DomainError) Error() string {
	if e.Cause == nil {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
}

func (e *DomainError) Unwrap() error { return e.Cause }

func NewError(code ErrorCode, message string) *DomainError {
	return &DomainError{Code: code, Message: message, Fields: map[string]string{}}
}

func FieldError(field, message string) *DomainError {
	return &DomainError{
		Code:    CodeInvalid,
		Message: "request validation failed",
		Fields:  map[string]string{field: message},
	}
}

func Wrap(code ErrorCode, message string, cause error) *DomainError {
	return &DomainError{Code: code, Message: message, Cause: cause, Fields: map[string]string{}}
}

func CodeOf(err error) ErrorCode {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Code
	}
	return CodeInternal
}
