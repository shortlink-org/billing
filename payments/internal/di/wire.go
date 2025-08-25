//go:generate go tool wire
//go:build wireinject

package di

import (
	"github.com/google/wire"

	"github.com/shortlink-org/billing/payments/internal/application/payments/usecase/create"
	"github.com/shortlink-org/billing/payments/internal/application/payments/usecase/refund"
)

// Wire provider sets for dependency injection
var (
	// InfrastructureSet provides all infrastructure dependencies
	InfrastructureSet = wire.NewSet(
		ProvidePaymentRepository,
		ProvidePaymentProvider,
	)

	// UsecaseSet provides all usecase handlers
	UsecaseSet = wire.NewSet(
		ProvideCreateHandler,
		ProvideRefundHandler,
	)

	// AllProviders combines all provider sets
	AllProviders = wire.NewSet(
		InfrastructureSet,
		UsecaseSet,
	)
)

// InitializeCreateHandler wires up the create payment usecase handler
func InitializeCreateHandler(cfg PaymentConfig) (*create.Handler, error) {
	wire.Build(AllProviders)
	return nil, nil
}

// InitializeRefundHandler wires up the refund payment usecase handler
func InitializeRefundHandler(cfg PaymentConfig) (*refund.Handler, error) {
	wire.Build(AllProviders)
	return nil, nil
}

// InitializePaymentUsecases wires up all payment usecase handlers
func InitializePaymentUsecases(cfg PaymentConfig) (*PaymentUsecases, error) {
	wire.Build(
		AllProviders,
		wire.Struct(new(PaymentUsecases), "*"),
	)
	return nil, nil
}
