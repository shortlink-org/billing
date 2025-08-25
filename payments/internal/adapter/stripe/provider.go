package stripeadp

import (
	"context"
	"strings"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/paymentintent"
	"github.com/stripe/stripe-go/v82/refund"

	"google.golang.org/genproto/googleapis/type/money"

	"github.com/shortlink-org/billing/payments/internal/application/payments/ports"
	"github.com/shortlink-org/billing/payments/internal/domain/payment/ledger"
)

type Config struct{ APIKey string }

type Provider struct{ c *stripe.Client }

func New(cfg Config) *Provider {
	return &Provider{c: stripe.NewClient(cfg.APIKey)}
}

func (p *Provider) CreatePayment(ctx context.Context, in ports.CreatePaymentIn) (ports.CreatePaymentOut, error) {
	minor, err := ledger.AmountToMinorUnits(in.Amount)
	if err != nil {
		return ports.CreatePaymentOut{}, err
	}

	currency := strings.ToLower(in.Currency)
	captureMethod := stripe.String(string(stripe.PaymentIntentCaptureMethodAutomatic))
	if in.CaptureManual {
		captureMethod = stripe.String(string(stripe.PaymentIntentCaptureMethodManual))
	}

	params := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(minor),
		Currency:      stripe.String(currency),
		CaptureMethod: captureMethod,
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
		Description: stripe.String(in.Description),
	}
	// Attach context properly.
	params.Context = ctx

	// Optional: strong idempotency key (highly recommended).
	params.SetIdempotencyKey(in.PaymentID.String())

	// Metadata (use AddMetadata on embedded stripe.Params).
	for k, v := range in.Metadata {
		params.AddMetadata(k, v)
	}

	if in.ReturnURL != "" {
		params.ReturnURL = stripe.String(in.ReturnURL)
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		return ports.CreatePaymentOut{}, err
	}

	out := ports.CreatePaymentOut{
		Provider:     ports.ProviderStripe,
		ProviderID:   pi.ID,
		ClientSecret: pi.ClientSecret, // return only to API caller, never to events
		Status:       mapPIStatus(pi),
	}

	switch out.Status {
	case ports.ProviderStatusRequiresCapture:
		out.Authorized = fromMinor(pi.Currency, pi.Amount)
	case ports.ProviderStatusSucceeded:
		out.Captured = fromMinor(pi.Currency, pi.AmountReceived)
	}

	return out, nil
}

func (p *Provider) RefundPayment(ctx context.Context, in ports.RefundPaymentIn) (ports.RefundPaymentOut, error) {
	minor, err := ledger.AmountToMinorUnits(in.Amount)
	if err != nil {
		return ports.RefundPaymentOut{}, err
	}

	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(in.ProviderID),
		Amount:        stripe.Int64(minor),
		Reason:        stripe.String(stripe.RefundReasonRequestedByCustomer), // default reason
	}
	// Attach context properly.
	params.Context = ctx

	// Optional: strong idempotency key (highly recommended).
	params.SetIdempotencyKey(in.PaymentID.String() + "_refund")

	// Metadata (use AddMetadata on embedded stripe.Params).
	for k, v := range in.Metadata {
		params.AddMetadata(k, v)
	}

	if in.Reason != "" {
		// Map common reasons to Stripe-specific reasons
		switch strings.ToLower(in.Reason) {
		case "duplicate", "duplicated":
			params.Reason = stripe.String(stripe.RefundReasonDuplicate)
		case "fraud", "fraudulent":
			params.Reason = stripe.String(stripe.RefundReasonFraudulent)
		default:
			params.Reason = stripe.String(stripe.RefundReasonRequestedByCustomer)
		}
	}

	r, err := refund.New(params)
	if err != nil {
		return ports.RefundPaymentOut{}, err
	}

	out := ports.RefundPaymentOut{
		Provider: ports.ProviderStripe,
		RefundID: r.ID,
		Status:   mapRefundStatus(r),
		Amount:   fromMinor(r.Currency, r.Amount),
	}

	return out, nil
}

func fromMinor(cur stripe.Currency, v int64) *money.Money {
	return ledger.MinorUnitsToAmount(string(cur), v)
}

func mapPIStatus(pi *stripe.PaymentIntent) ports.ProviderStatus {
	switch pi.Status {
	case stripe.PaymentIntentStatusRequiresAction:
		return ports.ProviderStatusRequiresAction
	case stripe.PaymentIntentStatusRequiresCapture:
		return ports.ProviderStatusRequiresCapture
	case stripe.PaymentIntentStatusSucceeded:
		return ports.ProviderStatusSucceeded
	case stripe.PaymentIntentStatusRequiresPaymentMethod,
		stripe.PaymentIntentStatusRequiresConfirmation,
		stripe.PaymentIntentStatusProcessing:
		return ports.ProviderStatusPending
	case stripe.PaymentIntentStatusCanceled:
		return ports.ProviderStatusCanceled
	default:
		return ports.ProviderStatusUnknown
	}
}

func mapRefundStatus(r *stripe.Refund) ports.ProviderStatus {
	switch r.Status {
	case stripe.RefundStatusSucceeded:
		return ports.ProviderStatusSucceeded
	case stripe.RefundStatusPending:
		return ports.ProviderStatusPending
	case stripe.RefundStatusFailed:
		return ports.ProviderStatusFailed
	case stripe.RefundStatusCanceled:
		return ports.ProviderStatusCanceled
	default:
		return ports.ProviderStatusUnknown
	}
}
