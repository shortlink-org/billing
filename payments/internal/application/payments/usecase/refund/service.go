package refund

import (
	"context"
	"fmt"

	"github.com/shortlink-org/billing/payments/internal/application/payments/ports"
	"github.com/shortlink-org/billing/payments/internal/application/payments/repository"
)

// Service provides the application service layer for refund operations.
// It handles DTOs and coordinates with the domain handler.
type Service struct {
	handler *Handler
}

// NewService creates a new refund service with the required dependencies.
func NewService(repo repository.PaymentRepository, provider ports.PaymentProvider) *Service {
	return &Service{
		handler: &Handler{
			Repo:     repo,
			Provider: provider,
		},
	}
}

// ProcessRefund handles the refund operation using DTOs.
// It validates the input DTO, converts it to domain objects, executes the business logic,
// and returns the result as a DTO.
func (s *Service) ProcessRefund(ctx context.Context, req RefundRequestDTO) (*RefundResponseDTO, error) {
	// Validate DTO (in a real application, you might use a validator like go-playground/validator)
	if err := s.validateRefundRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Convert DTO to domain command
	cmd, err := req.ToCommand()
	if err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Execute domain logic
	result, err := s.handler.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	// Convert domain result to DTO
	response := FromResult(result)
	return &response, nil
}

// validateRefundRequest performs basic validation on the refund request DTO.
func (s *Service) validateRefundRequest(req RefundRequestDTO) error {
	if req.PaymentID == "" {
		return fmt.Errorf("%w: payment ID is required", ErrInvalidRefundAmount)
	}

	if req.Reason == "" {
		return fmt.Errorf("%w: reason is required", ErrInvalidRefundAmount)
	}

	if req.Amount != nil {
		if req.Amount.CurrencyCode == "" {
			return fmt.Errorf("%w: currency code is required when amount is specified", ErrInvalidRefundAmount)
		}
		if len(req.Amount.CurrencyCode) != 3 {
			return fmt.Errorf("%w: currency code must be 3 characters", ErrInvalidRefundAmount)
		}
		if req.Amount.Units < 0 || (req.Amount.Units == 0 && req.Amount.Nanos <= 0) {
			return fmt.Errorf("%w: amount must be positive", ErrInvalidRefundAmount)
		}
		if req.Amount.Nanos < -999999999 || req.Amount.Nanos > 999999999 {
			return fmt.Errorf("%w: nanos must be between -999999999 and 999999999", ErrInvalidRefundAmount)
		}
	}

	return nil
}