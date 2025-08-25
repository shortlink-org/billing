//go:generate wire
//go:build wireinject

// The build tag makes sure the stub is not built in the final build.

/*
Wallet Service DI-package (Updated Example)
This demonstrates how to update the wallet service to use the new default services pattern
*/
package wallet_di

import (
	"context"

	"github.com/google/wire"
	"go.opentelemetry.io/otel/trace"

	"github.com/shortlink-org/shortlink/pkg/di"
	"github.com/shortlink-org/shortlink/pkg/di/pkg/autoMaxPro"
	"github.com/shortlink-org/shortlink/pkg/di/pkg/config"
	"github.com/shortlink-org/shortlink/pkg/di/pkg/profiling"
	"github.com/shortlink-org/shortlink/pkg/di/providers"
	"github.com/shortlink-org/shortlink/pkg/logger"
	"github.com/shortlink-org/shortlink/pkg/observability/metrics"
)

// WalletServiceUpdated demonstrates the updated structure with full default services
type WalletServiceUpdated struct {
	// Common services
	Log        logger.Logger
	Config     *config.Config
	AutoMaxPro autoMaxPro.AutoMaxPro
	Context    context.Context

	// Observability
	Tracer        trace.TracerProvider
	Metrics       *metrics.Monitoring
	PprofEndpoint profiling.PprofEndpoint

	// Wallet-specific services would go here
	// Repository    repository.Repository
	// PaymentClient payment.Client
}

// Approach 1: Using modular providers (gives you control over what to include)
var WalletSetModular = wire.NewSet(
	// Default services (pick what you need)
	providers.ProvideContext,
	providers.ProvideLogger,
	providers.ProvideConfig,
	providers.ProvideAutoMaxPro,
	providers.ProvideTracer,
	providers.ProvideMetrics,
	providers.ProvidePprofEndpoint,

	// Wallet-specific providers would go here
	// repository.New,
	// payment.NewClient,

	NewWalletServiceModular,
)

// Approach 2: Using consolidated DefaultSet (includes all default services)
var WalletSetConsolidated = wire.NewSet(
	di.DefaultSet, // All default services

	// Wallet-specific providers would go here
	// repository.New,
	// payment.NewClient,

	NewWalletServiceConsolidated,
)

// Approach 3: Minimal - only include what the original service used
var WalletSetMinimal = wire.NewSet(
	providers.ProvideLogger,

	// Wallet-specific providers
	// Add as needed

	NewWalletServiceMinimal,
)

// Constructor using individual dependencies (Approach 1)
func NewWalletServiceModular(
	ctx context.Context,
	log logger.Logger,
	config *config.Config,
	autoMaxPro autoMaxPro.AutoMaxPro,
	tracer trace.TracerProvider,
	metrics *metrics.Monitoring,
	pprofEndpoint profiling.PprofEndpoint,
	// Add wallet-specific dependencies here
	// repository repository.Repository,
	// paymentClient payment.Client,
) (*WalletServiceUpdated, error) {
	log.Info("Start wallet service (modular)")

	return &WalletServiceUpdated{
		Context:    ctx,
		Log:        log,
		Config:     config,
		AutoMaxPro: autoMaxPro,

		Tracer:        tracer,
		Metrics:       metrics,
		PprofEndpoint: pprofEndpoint,

		// Wallet-specific services
		// Repository:    repository,
		// PaymentClient: paymentClient,
	}, nil
}

// Constructor using DefaultServices struct (Approach 2)
func NewWalletServiceConsolidated(
	defaults *di.DefaultServices,
	// Add wallet-specific dependencies here
	// repository repository.Repository,
	// paymentClient payment.Client,
) (*WalletServiceUpdated, error) {
	defaults.Log.Info("Start wallet service (consolidated)")

	return &WalletServiceUpdated{
		Context:    defaults.Context,
		Log:        defaults.Log,
		Config:     defaults.Config,
		AutoMaxPro: defaults.AutoMaxPro,

		Tracer:        defaults.Tracer,
		Metrics:       defaults.Metrics,
		PprofEndpoint: defaults.PprofEndpoint,

		// Wallet-specific services
		// Repository:    repository,
		// PaymentClient: paymentClient,
	}, nil
}

// Constructor using minimal dependencies (Approach 3 - similar to original)
func NewWalletServiceMinimal(
	log logger.Logger,
	// Add other minimal dependencies as needed
) (*WalletServiceUpdated, error) {
	log.Info("Start wallet service (minimal)")

	return &WalletServiceUpdated{
		Log: log,
		// Only set the services you actually need
	}, nil
}

// Wire injector functions for each approach
func InitializeWalletServiceModular() (*WalletServiceUpdated, func(), error) {
	panic(wire.Build(WalletSetModular))
}

func InitializeWalletServiceConsolidated() (*WalletServiceUpdated, func(), error) {
	panic(wire.Build(WalletSetConsolidated))
}

func InitializeWalletServiceMinimal() (*WalletServiceUpdated, func(), error) {
	panic(wire.Build(WalletSetMinimal))
}