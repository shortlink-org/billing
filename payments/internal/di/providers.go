package di

import (
	stripeadp "github.com/shortlink-org/billing/payments/internal/adapter/stripe"
	"github.com/shortlink-org/billing/payments/internal/application/payments/ports"
	"github.com/shortlink-org/billing/payments/internal/application/payments/repository"
	"github.com/shortlink-org/billing/payments/internal/application/payments/repository/memory"
	"github.com/shortlink-org/billing/payments/internal/application/payments/usecase/create"
	"github.com/shortlink-org/billing/payments/internal/application/payments/usecase/refund"
)

// PaymentUsecases holds all payment-related usecases for dependency injection.
type PaymentUsecases struct {
	CreatePayment *create.Handler
	RefundPayment *refund.Handler
}

// PaymentServices holds all payment-related services for dependency injection.
type PaymentServices struct {
	RefundService *refund.Service
}

// PaymentConfig holds configuration for payment services.
type PaymentConfig struct {
	StripeAPIKey string
}

// NewPaymentUsecases creates and wires all payment usecases with their dependencies.
// This follows the dependency injection pattern and can be easily extended
// with other DI frameworks like Wire, Dig, or Fx.
func NewPaymentUsecases(cfg PaymentConfig) *PaymentUsecases {
	// Infrastructure layer - repositories
	paymentRepo := memory.New()

	// Infrastructure layer - external providers
	stripeProvider := stripeadp.New(stripeadp.Config{
		APIKey: cfg.StripeAPIKey,
	})

	// Application layer - usecases
	createHandler := &create.Handler{
		Repo:     paymentRepo,
		Provider: stripeProvider,
	}

	refundHandler := &refund.Handler{
		Repo:     paymentRepo,
		Provider: stripeProvider,
	}

	return &PaymentUsecases{
		CreatePayment: createHandler,
		RefundPayment: refundHandler,
	}
}

// Alternative: Using interfaces for better testability and flexibility
type PaymentDependencies struct {
	Repository repository.PaymentRepository
	Provider   ports.PaymentProvider
}

// NewPaymentUsecasesWithDeps creates usecases with pre-configured dependencies.
// This pattern is useful when using external DI containers.
func NewPaymentUsecasesWithDeps(deps PaymentDependencies) *PaymentUsecases {
	createHandler := &create.Handler{
		Repo:     deps.Repository,
		Provider: deps.Provider,
	}

	refundHandler := &refund.Handler{
		Repo:     deps.Repository,
		Provider: deps.Provider,
	}

	return &PaymentUsecases{
		CreatePayment: createHandler,
		RefundPayment: refundHandler,
	}
}

// Wire-style providers (for use with Google Wire)
// These functions would be used in a wire.go file for automatic DI generation

// ProvidePaymentRepository provides the payment repository implementation.
func ProvidePaymentRepository() repository.PaymentRepository {
	return memory.New()
}

// ProvidePaymentProvider provides the payment provider implementation.
func ProvidePaymentProvider(cfg PaymentConfig) ports.PaymentProvider {
	return stripeadp.New(stripeadp.Config{
		APIKey: cfg.StripeAPIKey,
	})
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

// ProvideRefundService provides the refund payment service.
func ProvideRefundService(
	repo repository.PaymentRepository,
	provider ports.PaymentProvider,
) *refund.Service {
	return refund.NewService(repo, provider)
}

// NewPaymentServices creates and wires all payment services with their dependencies.
func NewPaymentServices(cfg PaymentConfig) *PaymentServices {
	// Infrastructure layer - repositories
	paymentRepo := memory.New()

	// Infrastructure layer - external providers
	stripeProvider := stripeadp.New(stripeadp.Config{
		APIKey: cfg.StripeAPIKey,
	})

	// Application layer - services
	refundService := refund.NewService(paymentRepo, stripeProvider)

	return &PaymentServices{
		RefundService: refundService,
	}
}