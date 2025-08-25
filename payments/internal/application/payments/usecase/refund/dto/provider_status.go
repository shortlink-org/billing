package dto

import (
	"github.com/stripe/stripe-go/v82"
	"github.com/shortlink-org/billing/payments/internal/application/payments/ports"
)

// ProviderStatusDTO represents provider status in a DTO-friendly format.
type ProviderStatusDTO string

const (
	ProviderStatusUnknown        ProviderStatusDTO = "unknown"
	ProviderStatusRequiresAction ProviderStatusDTO = "requires_action"
	ProviderStatusRequiresCapture ProviderStatusDTO = "requires_capture"
	ProviderStatusSucceeded      ProviderStatusDTO = "succeeded"
	ProviderStatusPending        ProviderStatusDTO = "pending"
	ProviderStatusCanceled       ProviderStatusDTO = "canceled"
	ProviderStatusFailed         ProviderStatusDTO = "failed"
)

// MapRefundStatus maps Stripe refund status to provider status DTO.
func MapRefundStatus(r *stripe.Refund) ProviderStatusDTO {
	switch r.Status {
	case stripe.RefundStatusSucceeded:
		return ProviderStatusSucceeded
	case stripe.RefundStatusPending:
		return ProviderStatusPending
	case stripe.RefundStatusFailed:
		return ProviderStatusFailed
	case stripe.RefundStatusCanceled:
		return ProviderStatusCanceled
	default:
		return ProviderStatusUnknown
	}
}

// ToProviderStatus converts ProviderStatusDTO to internal provider status.
func (dto ProviderStatusDTO) ToProviderStatus() ports.ProviderStatus {
	switch dto {
	case ProviderStatusRequiresAction:
		return ports.ProviderStatusRequiresAction
	case ProviderStatusRequiresCapture:
		return ports.ProviderStatusRequiresCapture
	case ProviderStatusSucceeded:
		return ports.ProviderStatusSucceeded
	case ProviderStatusPending:
		return ports.ProviderStatusPending
	case ProviderStatusCanceled:
		return ports.ProviderStatusCanceled
	case ProviderStatusFailed:
		return ports.ProviderStatusFailed
	default:
		return ports.ProviderStatusUnknown
	}
}

// FromProviderStatus converts internal provider status to ProviderStatusDTO.
func FromProviderStatus(status ports.ProviderStatus) ProviderStatusDTO {
	switch status {
	case ports.ProviderStatusRequiresAction:
		return ProviderStatusRequiresAction
	case ports.ProviderStatusRequiresCapture:
		return ProviderStatusRequiresCapture
	case ports.ProviderStatusSucceeded:
		return ProviderStatusSucceeded
	case ports.ProviderStatusPending:
		return ProviderStatusPending
	case ports.ProviderStatusCanceled:
		return ProviderStatusCanceled
	case ports.ProviderStatusFailed:
		return ProviderStatusFailed
	default:
		return ProviderStatusUnknown
	}
}