package app

import (
	"errors"
	"fmt"
	"net/http"
)

// CodedError implements app code helper. it shouldn't be used directly.
type CodedError struct {
	Err        error
	AppCode    int
	StatusCode int
}

// Error returns internal error or by status code.
func (e *CodedError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}

	if s := http.StatusText(e.StatusCode); s != "" {
		return s
	}

	return "unknown coded error"
}

// WithCode adds code to error.
func WithCode(err error, code int) error {
	if err == nil {
		return nil
	}

	codeError, ok := err.(*CodedError)
	if ok {
		if codeError == nil {
			return nil
		}

		codeError.AppCode = code
		return codeError
	}

	return &CodedError{
		Err:        err,
		StatusCode: http.StatusInternalServerError,
		AppCode:    code,
	}
}

// AuthError generates auth error.
func AuthError() error {
	return &CodedError{
		StatusCode: http.StatusUnauthorized}
}

// Unauthorized generates an unauthorized error with a custom message.
func Unauthorized(err error) error {
	return &CodedError{
		Err:        err,
		StatusCode: http.StatusUnauthorized,
		AppCode:    CodeUnauthorized,
	}
}

// Forbidden generates a forbidden error with a custom message.
func Forbidden(err error) error {
	return &CodedError{
		Err:        err,
		StatusCode: http.StatusForbidden,
		AppCode:    CodeForbidden,
	}
}

// ParamError generates BadParamError.
func ParamError(err error) error {
	if err != nil {
		return &CodedError{
			Err:        err,
			StatusCode: http.StatusBadRequest,
			AppCode:    CodeBadParam,
		}
	}

	return nil
}

// NotFoundError indicates resource not found.
func NotFoundError() error {
	return &CodedError{
		Err:        errors.New("error not found"),
		StatusCode: http.StatusNotFound,
		AppCode:    CodeNotFound,
	}
}

// InternalError creates an internal server error.
func InternalError(err error) *CodedError {
	return &CodedError{
		Err:        fmt.Errorf("internal server error: %w", err),
		StatusCode: http.StatusInternalServerError,
		AppCode:    CodeInternalError,
	}
}
