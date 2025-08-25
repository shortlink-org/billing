package ports

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/genproto/googleapis/type/money"
)

// Provider identifier (per ADR naming).
type Provider string

const (
	ProviderStripe Provider = "stripe"
)

// Normalized provider status after CreatePayment.
type ProviderStatus int

const (
	ProviderStatusUnknown        ProviderStatus = iota
	ProviderStatusRequiresAction                // 3DS/SCA
	ProviderStatusRequiresCapture
	ProviderStatusSucceeded
	ProviderStatusPending // requires_payment_method / requires_confirmation / processing
	ProviderStatusCanceled
	ProviderStatusFailed
)

type CreatePaymentIn struct {
	PaymentID     uuid.UUID
	InvoiceID     uuid.UUID
	Amount        *money.Money
	Currency      string // ISO-4217 (dup for convenience)
	CaptureManual bool   // true → manual capture, false → auto
	Description   string
	Metadata      map[string]string
	ReturnURL     string
}

type CreatePaymentOut struct {
	Provider     Provider
	ProviderID   string // e.g., Stripe PaymentIntent ID
	ClientSecret string // return to API only (never to events)
	Status       ProviderStatus

	Authorized *money.Money // set if provider holds funds (requires_capture)
	Captured   *money.Money // set if provider captured (succeeded)
}

type PaymentProvider interface {
	CreatePayment(ctx context.Context, in CreatePaymentIn) (CreatePaymentOut, error)
}
