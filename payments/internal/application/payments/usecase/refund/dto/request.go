package dto

import (
	"github.com/google/uuid"
	"google.golang.org/genproto/googleapis/type/money"
)

// RefundRequestDTO represents the input data for refunding a payment.
type RefundRequestDTO struct {
	PaymentID string            `json:"payment_id" validate:"required,uuid"`
	Amount    *MoneyDTO         `json:"amount,omitempty"` // nil for full refund
	Reason    string            `json:"reason" validate:"required"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// ToCommand converts RefundRequestDTO to the domain Command.
func (dto RefundRequestDTO) ToCommand() (Command, error) {
	paymentID, err := uuid.Parse(dto.PaymentID)
	if err != nil {
		return Command{}, err
	}

	var amount *money.Money
	if dto.Amount != nil {
		amount = dto.Amount.ToMoney()
	}

	return Command{
		PaymentID: paymentID,
		Amount:    amount,
		Reason:    dto.Reason,
		Metadata:  dto.Metadata,
	}, nil
}

// Command represents the internal domain command for refund operations.
// This is temporarily here to avoid circular imports, but should be moved to the handler package.
type Command struct {
	PaymentID uuid.UUID
	Amount    *money.Money
	Reason    string
	Metadata  map[string]string
}
