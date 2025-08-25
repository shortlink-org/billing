//go:generate go tool wire
//go:build wireinject

/*
Example Service DI-package
This demonstrates how to use the default services pattern
*/
package di

import (
	"context"

	"github.com/google/wire"
	"go.opentelemetry.io/otel/trace"

	"github.com/shortlink-org/shortlink/pkg/di/pkg/autoMaxPro"
	"github.com/shortlink-org/shortlink/pkg/di/pkg/config"
	"github.com/shortlink-org/shortlink/pkg/di/pkg/profiling"
	"github.com/shortlink-org/shortlink/pkg/di/providers"
	"github.com/shortlink-org/shortlink/pkg/logger"
	"github.com/shortlink-org/shortlink/pkg/observability/metrics"
)

// ExampleService demonstrates the service structure with default services
type ExampleService struct {
	// Common services (from DefaultSet)
	Log        logger.Logger
	Config     *config.Config
	AutoMaxPro autoMaxPro.AutoMaxPro
	Context    context.Context

	// Observability services (from DefaultSet)
	Tracer        trace.TracerProvider
	Metrics       *metrics.Monitoring
	PprofEndpoint profiling.PprofEndpoint

	// Service-specific dependencies would go here
	// For example:
	// Repository    repository.Repository
	// APIClient     client.APIClient
	// Cache         cache.Cache
}

// ExampleSet combines default services with service-specific providers
var ExampleSet = wire.NewSet(
	// Include all default services
	providers.ProvideContext,
	providers.ProvideLogger,
	providers.ProvideConfig,
	providers.ProvideAutoMaxPro,
	providers.ProvideTracer,
	providers.ProvideMetrics,
	providers.ProvidePprofEndpoint,

	// Add service-specific providers here
	// For example:
	// repository.New,
	// client.New,
	// cache.New,

	// Service constructor
	NewExampleService,
)

// NewExampleService creates a new ExampleService with all dependencies
func NewExampleService(
	// Common services
	ctx context.Context,
	log logger.Logger,
	config *config.Config,
	autoMaxPro autoMaxPro.AutoMaxPro,

	// Observability services
	tracer trace.TracerProvider,
	metrics *metrics.Monitoring,
	pprofEndpoint profiling.PprofEndpoint,

	// Service-specific dependencies would be added here
	// For example:
	// repository repository.Repository,
	// apiClient client.APIClient,
	// cache cache.Cache,
) (*ExampleService, error) {
	log.Info("Initializing ExampleService")

	return &ExampleService{
		// Common services
		Context:    ctx,
		Log:        log,
		Config:     config,
		AutoMaxPro: autoMaxPro,

		// Observability services
		Tracer:        tracer,
		Metrics:       metrics,
		PprofEndpoint: pprofEndpoint,

		// Service-specific dependencies
		// Repository: repository,
		// APIClient:  apiClient,
		// Cache:      cache,
	}, nil
}

// InitializeExampleService wires up the example service
func InitializeExampleService() (*ExampleService, func(), error) {
	panic(wire.Build(ExampleSet))
}