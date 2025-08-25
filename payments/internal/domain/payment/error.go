package payment

import "errors"

var (
	ErrInvalidArgs         = errors.New("payment: invalid arguments")
	ErrInvalidTransition   = errors.New("payment: invalid transition")
	ErrTerminalState       = errors.New("payment: terminal state")
	ErrPolicyCaptureMode   = errors.New("payment: capture not allowed from CREATED in MANUAL mode")
	ErrUnsupportedCurrency = errors.New("payment: unsupported currency")
	ErrInvariantViolation  = errors.New("payment: invariants violated")
	ErrBadPaymentID        = errors.New("payment: invalid meta.payment_id bytes")
	ErrBadInvoiceID        = errors.New("payment: invalid invoice_id bytes")
)
