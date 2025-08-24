package payment

import (
	"errors"
)

var (
	ErrInvalidArgs       = errors.New("payment: invalid arguments")
	ErrInvalidTransition = errors.New("payment: invalid transition")
	ErrTerminalState     = errors.New("payment: terminal state")
	ErrPolicyCaptureMode = errors.New("payment: capture not allowed in MANUAL mode from CREATED")
	ErrVersionConflict   = errors.New("payment: version conflict")
	ErrNotFound          = errors.New("payment: not found")
)
