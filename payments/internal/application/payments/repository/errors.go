package repository

import "errors"

// ErrNotFound is returned when a payment stream does not exist.
var ErrNotFound = errors.New("payment repository: not found")
