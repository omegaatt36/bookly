package app

import (
	"errors"
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
	}
}
