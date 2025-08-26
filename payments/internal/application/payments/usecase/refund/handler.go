package refund

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/samber/lo"

	"github.com/shortlink-org/billing/payments/internal/application/payments/ports"
	"github.com/shortlink-org/billing/payments/internal/application/payments/repository"
	"github.com/shortlink-org/billing/payments/internal/application/payments/usecase/refund/dto"
	eventv1 "github.com/shortlink-org/billing/payments/internal/domain/event/v1"
	flowv1 "github.com/shortlink-org/billing/payments/internal/domain/flow/v1"
	"github.com/shortlink-org/billing/payments/internal/domain/payment/ledger"

	"google.golang.org/genproto/googleapis/type/money"
)

// Helper functions for money validation
func isZero(m *money.Money) bool {
	return m == nil || (m.Units == 0 && m.Nanos == 0)
}

func isNegative(m *money.Money) bool {
	if m == nil {
		return false
	}
	return m.Units < 0 || (m.Units == 0 && m.Nanos < 0)
}

// Result is returned after successful refund initiation.
type Result struct {
	PaymentID     uuid.UUID
	RefundID      string
	RefundAmount  *money.Money
	TotalRefunded *money.Money
	IsFullRefund  bool
	State         flowv1.PaymentFlow
	Version       uint64
}

// Handler orchestrates payment refunds.
type Handler struct {
	Repo     repository.PaymentRepository
	Provider ports.PaymentProvider
}

func (h *Handler) Handle(ctx context.Context, cmd dto.Command) (*dto.Result, error) {
	// Validate input
	if cmd.PaymentID == uuid.Nil {
		return nil, fmt.Errorf("%w: payment ID is required", ErrInvalidRefundAmount)
	}

	// Load the payment aggregate
	agg, err := h.Repo.Load(ctx, cmd.PaymentID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("%w: %s", ErrPaymentNotFound, cmd.PaymentID)
		}
		return nil, fmt.Errorf("load payment: %w", err)
	}

	// Validate payment is refundable
	if agg.State() != flowv1.PaymentFlow_PAYMENT_FLOW_PAID && agg.State() != flowv1.PaymentFlow_PAYMENT_FLOW_REFUNDED {
		return nil, fmt.Errorf("%w: payment state is %v", ErrPaymentNotRefundable, agg.State())
	}

	// Determine refund amount
	refundAmount := cmd.Amount
	if refundAmount == nil {
		if agg.Ledger.Captured == nil {
			return nil, fmt.Errorf("%w: no captured amount to refund", ErrPaymentNotRefundable)
		}

		totalRefunded := agg.Ledger.TotalRefunded
		if totalRefunded == nil {
			totalRefunded = ledger.Zero(agg.Ledger.Captured.GetCurrencyCode())
		}

		remaining, err := ledger.Sub(agg.Ledger.Captured, totalRefunded)
		if err != nil {
			return nil, fmt.Errorf("calculate refundable amount: %w", err)
		}

		if isZero(remaining) {
			return nil, fmt.Errorf("%w: payment already fully refunded", ErrInvalidRefundAmount)
		}

		refundAmount = remaining
	}

	// Validate refund amount
	if refundAmount == nil || isZero(refundAmount) || isNegative(refundAmount) {
		return nil, fmt.Errorf("%w: amount must be positive", ErrInvalidRefundAmount)
	}

	// TODO: extract providerID from aggregate/metadata when available
	providerID := ""

	providerIn := ports.RefundPaymentIn{
		PaymentID:  cmd.PaymentID,
		ProviderID: providerID,
		Amount:     refundAmount,
		Currency:   refundAmount.GetCurrencyCode(),
		Reason:     cmd.Reason,
		Metadata: lo.Assign(cmd.Metadata, map[string]string{
			"payment_id":    cmd.PaymentID.String(),
			"refund_reason": cmd.Reason,
		}),
	}

	providerOut, err := h.Provider.RefundPayment(ctx, providerIn)
	if err != nil {
		// провайдер/интеграционная ошибка -> NETWORK_ERROR
		agg.RefundFailed(ctx, eventv1.FailureReason_FAILURE_REASON_NETWORK_ERROR)

		if saveErr := h.Repo.Save(ctx, agg, agg.Version()); saveErr != nil {
			return nil, fmt.Errorf("save refund failure: %w (original error: %v)", saveErr, err)
		}
		return nil, fmt.Errorf("provider refund failed: %w", err)
	}

	// Apply refund to domain aggregate
	actualRefundAmount := lo.Ternary(providerOut.Amount != nil, providerOut.Amount, refundAmount)
	isFullRefund, err := agg.Refund(ctx, actualRefundAmount)
	if err != nil {
		return nil, fmt.Errorf("apply refund to aggregate: %w", err)
	}

	if err := agg.Invariants(); err != nil {
		return nil, fmt.Errorf("domain invariants violated: %w", err)
	}

	if err := h.Repo.Save(ctx, agg, agg.Version()-1); err != nil {
		return nil, fmt.Errorf("save refunded payment: %w", err)
	}

	return &dto.Result{
		PaymentID:     cmd.PaymentID,
		RefundID:      providerOut.RefundID,
		RefundAmount:  actualRefundAmount,
		TotalRefunded: agg.Ledger.TotalRefunded,
		IsFullRefund:  isFullRefund,
		State:         agg.State(),
		Version:       agg.Version(),
	}, nil
}
