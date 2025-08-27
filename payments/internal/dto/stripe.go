package dto

import (
	"github.com/stripe/stripe-go/v82"
	"github.com/shortlink-org/billing/payments/internal/application/payments/ports"
)

// MapPIStatus maps Stripe PaymentIntent status to provider status.
func MapPIStatus(pi *stripe.PaymentIntent) ports.ProviderStatus {
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

// MapRefundStatus maps Stripe refund status to provider status.
func MapRefundStatus(r *stripe.Refund) ports.ProviderStatus {
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