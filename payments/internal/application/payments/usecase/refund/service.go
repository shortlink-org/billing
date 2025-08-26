package refund

import (
	"context"

	"github.com/shortlink-org/billing/payments/internal/application/payments/ports"
	"github.com/shortlink-org/billing/payments/internal/application/payments/repository"
	"github.com/shortlink-org/billing/payments/internal/application/payments/usecase/refund/dto"
)

// Service provides the application service layer for refund operations.
type Service struct {
	handler *Handler
}

// NewService constructs a new refund service with the required dependencies.
func NewService(repo repository.PaymentRepository, provider ports.PaymentProvider) *Service {
	return &Service{
		handler: &Handler{
			Repo:     repo,
			Provider: provider,
		},
	}
}

// ProcessRefund validates the request DTO, converts it to a domain command,
// executes the refund business logic, and maps the result back to a DTO.
func (s *Service) ProcessRefund(ctx context.Context, req dto.RefundRequestDTO) (*dto.RefundResponseDTO, error) {
	if err := s.validateRefundRequest(req); err != nil {
		return nil, err
	}

	cmd, err := req.ToCommand()
	if err != nil {
		return nil, ErrInvalidRefundAmount
	}

	result, err := s.handler.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	resp := dto.FromResult(result)
	return &resp, nil
}

// validateRefundRequest performs basic validation on the refund request DTO.
func (s *Service) validateRefundRequest(req dto.RefundRequestDTO) error {
	if req.PaymentID == "" {
		return ErrInvalidRefundAmount
	}
	if req.Reason == "" {
		return ErrInvalidRefundAmount
	}
	if req.Amount != nil {
		if req.Amount.CurrencyCode == "" {
			return ErrInvalidRefundAmount
		}
		if len(req.Amount.CurrencyCode) != 3 {
			return ErrInvalidRefundAmount
		}
		if req.Amount.Units < 0 || (req.Amount.Units == 0 && req.Amount.Nanos <= 0) {
			return ErrInvalidRefundAmount
		}
		if req.Amount.Nanos < -999_999_999 || req.Amount.Nanos > 999_999_999 {
			return ErrInvalidRefundAmount
		}
	}
	return nil
}
