package domain

import "errors"

// ErrNotFound indicates that a requested resource was not found.
var ErrNotFound = errors.New("resource not found")
