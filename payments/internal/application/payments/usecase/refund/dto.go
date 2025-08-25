package refund

import (
	"github.com/google/uuid"
	flowv1 "github.com/shortlink-org/billing/payments/internal/domain/flow/v1"
	"google.golang.org/genproto/googleapis/type/money"
)

// RefundRequestDTO represents the input data for refunding a payment.
type RefundRequestDTO struct {
	PaymentID    string            `json:"payment_id" validate:"required,uuid"`
	Amount       *MoneyDTO         `json:"amount,omitempty"` // nil for full refund
	Reason       string            `json:"reason" validate:"required"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

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

// MoneyDTO represents a monetary amount.
type MoneyDTO struct {
	CurrencyCode string `json:"currency_code" validate:"required,len=3"`
	Units        int64  `json:"units"`
	Nanos        int32  `json:"nanos" validate:"gte=-999999999,lte=999999999"`
}

// ToCommand converts RefundRequestDTO to internal Command.
func (dto RefundRequestDTO) ToCommand() (Command, error) {
	paymentID, err := uuid.Parse(dto.PaymentID)
	if err != nil {
		return Command{}, err
	}

	var amount *MoneyDTO
	if dto.Amount != nil {
		amount = dto.Amount
	}

	return Command{
		PaymentID: paymentID,
		Amount:    amount.ToMoney(),
		Reason:    dto.Reason,
		Metadata:  dto.Metadata,
	}, nil
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

// ToMoney converts MoneyDTO to domain money type.
func (dto *MoneyDTO) ToMoney() *money.Money {
	if dto == nil {
		return nil
	}
	return &money.Money{
		CurrencyCode: dto.CurrencyCode,
		Units:        dto.Units,
		Nanos:        dto.Nanos,
	}
}

// FromMoney converts domain money type to MoneyDTO.
func FromMoney(m *money.Money) *MoneyDTO {
	if m == nil {
		return nil
	}
	return &MoneyDTO{
		CurrencyCode: m.GetCurrencyCode(),
		Units:        m.GetUnits(),
		Nanos:        m.GetNanos(),
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