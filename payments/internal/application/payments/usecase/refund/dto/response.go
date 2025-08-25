package dto

import (
	"github.com/google/uuid"
	flowv1 "github.com/shortlink-org/billing/payments/internal/domain/flow/v1"
	"google.golang.org/genproto/googleapis/type/money"
)

// RefundResponseDTO represents the output data after refund processing.
type RefundResponseDTO struct {
	PaymentID     string    `json:"payment_id"`
	RefundID      string    `json:"refund_id"`
	RefundAmount  *MoneyDTO `json:"refund_amount"`
	TotalRefunded *MoneyDTO `json:"total_refunded"`
	IsFullRefund  bool      `json:"is_full_refund"`
	State         string    `json:"state"`
	Version       uint64    `json:"version"`
}

// Result represents the internal domain result for refund operations.
// This is temporarily here to avoid circular imports, but should be moved to the handler package.
type Result struct {
	PaymentID     uuid.UUID
	RefundID      string
	RefundAmount  *money.Money
	TotalRefunded *money.Money
	IsFullRefund  bool
	State         flowv1.PaymentFlow
	Version       uint64
}

// FromResult converts internal Result to RefundResponseDTO.
func FromResult(result *Result) RefundResponseDTO {
	return RefundResponseDTO{
		PaymentID:     result.PaymentID.String(),
		RefundID:      result.RefundID,
		RefundAmount:  FromMoney(result.RefundAmount),
		TotalRefunded: FromMoney(result.TotalRefunded),
		IsFullRefund:  result.IsFullRefund,
		State:         mapFlowToString(result.State),
		Version:       result.Version,
	}
}

// mapFlowToString converts payment flow enum to string representation.
func mapFlowToString(flow flowv1.PaymentFlow) string {
	switch flow {
	case flowv1.PaymentFlow_PAYMENT_FLOW_CREATED:
		return "created"
	case flowv1.PaymentFlow_PAYMENT_FLOW_WAITING_FOR_CONFIRMATION:
		return "waiting_for_confirmation"
	case flowv1.PaymentFlow_PAYMENT_FLOW_AUTHORIZED:
		return "authorized"
	case flowv1.PaymentFlow_PAYMENT_FLOW_PAID:
		return "paid"
	case flowv1.PaymentFlow_PAYMENT_FLOW_REFUNDED:
		return "refunded"
	case flowv1.PaymentFlow_PAYMENT_FLOW_CANCELED:
		return "canceled"
	case flowv1.PaymentFlow_PAYMENT_FLOW_FAILED:
		return "failed"
	default:
		return "unknown"
	}
}