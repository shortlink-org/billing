package create

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/samber/lo"

	"github.com/shortlink-org/billing/payments/internal/application/payments/ports"
	"github.com/shortlink-org/billing/payments/internal/application/payments/repository"
	eventv1 "github.com/shortlink-org/billing/payments/internal/domain/event/v1"
	flowv1 "github.com/shortlink-org/billing/payments/internal/domain/flow/v1"
	"github.com/shortlink-org/billing/payments/internal/domain/payment"
	"github.com/shortlink-org/billing/payments/internal/domain/payment/ledger"

	"google.golang.org/genproto/googleapis/type/money"
)

// Command contains input data for creating a payment.
type Command struct {
	PaymentID   uuid.UUID // uuid.Nil → generate v7 in domain
	InvoiceID   uuid.UUID
	Amount      *money.Money
	Kind        eventv1.PaymentKind
	Mode        eventv1.CaptureMode
	Description string
	Metadata    map[string]string
	ReturnURL   string
}

// Result is returned after successful payment creation.
type Result struct {
	ID           uuid.UUID
	State        flowv1.PaymentFlow
	Version      uint64
	Provider     ports.Provider
	ProviderID   string
	ClientSecret string
}

// Handler orchestrates payment creation.
type Handler struct {
	Repo     repository.PaymentRepository
	Provider ports.PaymentProvider
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {
	agg, err := payment.New(cmd.PaymentID, cmd.InvoiceID, cmd.Amount, cmd.Kind, cmd.Mode)
	if err != nil {
		return nil, fmt.Errorf("create aggregate: %w", err)
	}

	// default metadata always overrides user metadata
	defaultMeta := map[string]string{
		"payment_id": agg.ID().String(),
		"invoice_id": agg.InvoiceID().String(),
		"kind":       cmd.Kind.String(),
		"mode":       cmd.Mode.String(),
	}
	meta := lo.Assign(cmd.Metadata, defaultMeta)

	out, err := h.Provider.CreatePayment(ctx, ports.CreatePaymentIn{
		PaymentID:     agg.ID(),
		InvoiceID:     agg.InvoiceID(),
		Amount:        cmd.Amount,
		Currency:      cmd.Amount.GetCurrencyCode(),
		CaptureManual: cmd.Mode == eventv1.CaptureMode_CAPTURE_MODE_MANUAL,
		Description:   cmd.Description,
		Metadata:      meta,
		ReturnURL:     cmd.ReturnURL,
	})
	if err != nil {
		return nil, fmt.Errorf("provider create: %w", err)
	}

	switch out.Status {
	case ports.ProviderStatusRequiresAction:
		if err := agg.RequireSCA(ctx); err != nil {
			return nil, err
		}
	case ports.ProviderStatusRequiresCapture:
		amt := lo.Ternary(out.Authorized != nil, out.Authorized, ledger.Clone(cmd.Amount))
		if err := agg.Authorize(ctx, amt); err != nil {
			return nil, err
		}
	case ports.ProviderStatusSucceeded:
		amt := lo.Ternary(out.Captured != nil, out.Captured, ledger.Clone(cmd.Amount))
		if err := agg.Capture(ctx, amt); err != nil {
			return nil, err
		}
	case ports.ProviderStatusPending:
		// no-op
	case ports.ProviderStatusCanceled:
		if err := agg.Cancel(ctx, eventv1.CancelReason_CANCEL_REASON_SYSTEM); err != nil {
			return nil, err
		}
	case ports.ProviderStatusFailed:
		if err := agg.Fail(ctx, eventv1.FailureReason_FAILURE_REASON_PROVIDER_ERROR); err != nil {
			return nil, err
		}
	default:
		// no-op
	}

	if err := agg.Invariants(); err != nil {
		return nil, fmt.Errorf("invariants: %w", err)
	}

	if err := h.Repo.Save(ctx, agg, 0); err != nil {
		return nil, fmt.Errorf("save: %w", err)
	}

	return &Result{
		ID:           agg.ID(),
		State:        agg.State(),
		Version:      agg.Version(),
		Provider:     out.Provider,
		ProviderID:   out.ProviderID,
		ClientSecret: out.ClientSecret,
	}, nil
}
