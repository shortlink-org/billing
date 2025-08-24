package ledger

import (
	"errors"
)

var (
	ErrNilAmount            = errors.New("ledger: amount is nil")
	ErrNilMoney             = errors.New("money: nil")
	ErrNonPositiveAmount    = errors.New("money: amount must be positive")
	ErrCurrencyMismatch     = errors.New("money: currency mismatch")
	ErrAuthorizeExceeds     = errors.New("authorize: would exceed amount")
	ErrCaptureExceedsLimit  = errors.New("capture: would exceed limit")
	ErrRefundWithoutCapture = errors.New("refund: nothing captured")
	ErrRefundExceeds        = errors.New("refund: would exceed captured")
)
