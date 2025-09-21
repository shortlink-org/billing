package di

import (
	"fmt"

	stripeadp "github.com/shortlink-org/billing/payments/internal/adapter/stripe"
	tinkoffadp "github.com/shortlink-org/billing/payments/internal/adapter/tinkoff"
	"github.com/shortlink-org/billing/payments/internal/application/payments/ports"
	"github.com/shortlink-org/billing/payments/internal/application/payments/repository"
	"github.com/shortlink-org/billing/payments/internal/application/payments/repository/memory"
	"github.com/shortlink-org/billing/payments/internal/application/payments/usecase/create"
	"github.com/shortlink-org/billing/payments/internal/application/payments/usecase/refund"
	"github.com/spf13/viper"
)

// ProvidePaymentRepository provides the payment repository implementation.
func ProvidePaymentRepository() repository.PaymentRepository {
	return memory.New()
}

// ProvidePaymentProvider provides the payment provider implementation.
// The provider is selected based on the PAYMENT_PROVIDER environment variable.
// Supported values: "stripe" (default), "tinkoff"
func ProvidePaymentProvider() (ports.PaymentProvider, error) {
	viper.AutomaticEnv()
	provider := viper.GetString("PAYMENT_PROVIDER")

	switch provider {
	case "tinkoff":
		return tinkoffadp.New()
	case "stripe", "":
		return stripeadp.New()
	default:
		return nil, fmt.Errorf("unsupported payment provider: %s", provider)
	}
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
