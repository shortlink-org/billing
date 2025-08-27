package stripeadp

import (
	"context"
	"strings"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/paymentintent"

	"github.com/shortlink-org/billing/payments/internal/application/payments/ports"
	"github.com/shortlink-org/billing/payments/internal/domain/payment/ledger"
	"github.com/shortlink-org/billing/payments/internal/dto"
)

// CreatePayment creates a new payment intent through Stripe.
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
		Status:       dto.MapPIStatus(pi),
	}

	switch out.Status {
	case ports.ProviderStatusRequiresCapture:
		out.Authorized = dto.FromMinor(pi.Currency, pi.Amount)
	case ports.ProviderStatusSucceeded:
		out.Captured = dto.FromMinor(pi.Currency, pi.AmountReceived)
	}

	return out, nil
}

