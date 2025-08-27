package stripeadp

import (
	"context"
	"strings"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/refund"

	"github.com/shortlink-org/billing/payments/internal/application/payments/ports"
	"github.com/shortlink-org/billing/payments/internal/domain/payment/ledger"
	"github.com/shortlink-org/billing/payments/internal/dto"
)

// RefundPayment creates a refund for a payment through Stripe.
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
		Status:   dto.MapRefundStatus(r),
		Amount:   dto.FromMinor(r.Currency, r.Amount),
	}

	return out, nil
}

