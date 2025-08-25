package refund

import "errors"

var (
	// ErrPaymentNotFound is returned when the payment to refund is not found.
	ErrPaymentNotFound = errors.New("refund: payment not found")
	// ErrInvalidRefundAmount is returned when the refund amount is invalid.
	ErrInvalidRefundAmount = errors.New("refund: invalid refund amount")
	// ErrPaymentNotRefundable is returned when the payment is not in a refundable state.
	ErrPaymentNotRefundable = errors.New("refund: payment is not refundable")
)