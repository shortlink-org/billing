package di

import (
	stripeadp "github.com/shortlink-org/billing/payments/internal/adapter/stripe"
	"github.com/shortlink-org/billing/payments/internal/application/payments/ports"
	"github.com/shortlink-org/billing/payments/internal/application/payments/repository"
	"github.com/shortlink-org/billing/payments/internal/application/payments/repository/memory"
	"github.com/shortlink-org/billing/payments/internal/application/payments/usecase/create"
	"github.com/shortlink-org/billing/payments/internal/application/payments/usecase/refund"
)

// ProvidePaymentRepository provides the payment repository implementation.
func ProvidePaymentRepository() repository.PaymentRepository {
	return memory.New()
}

// ProvidePaymentProvider provides the payment provider implementation.
func ProvidePaymentProvider() (ports.PaymentProvider, error) {
	return stripeadp.New()
}

// ProvideCreateHandler provides the create payment usecase handler.
func ProvideCreateHandler(
	repo repository.PaymentRepository,
	provider ports.PaymentProvider,
) *create.Handler {
	return &create.Handler{
		Repo:     repo,
		Provider: provider,
	}
}

// ProvideRefundHandler provides the refund payment usecase handler.
func ProvideRefundHandler(
	repo repository.PaymentRepository,
	provider ports.PaymentProvider,
) *refund.Handler {
	return &refund.Handler{
		Repo:     repo,
		Provider: provider,
	}
}
