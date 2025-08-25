//go:generate go tool wire
//go:build wireinject

/*
Auth Service DI-package (Updated Example)
This shows how to integrate the auth service with default services
*/
package di

import (
	"context"

	"github.com/authzed/authzed-go/v1"
	"github.com/google/wire"
	"go.opentelemetry.io/otel/trace"

	"github.com/shortlink-org/shortlink/pkg/di/pkg/autoMaxPro"
	"github.com/shortlink-org/shortlink/pkg/di/pkg/config"
	"github.com/shortlink-org/shortlink/pkg/di/pkg/profiling"
	"github.com/shortlink-org/shortlink/pkg/di/providers"
	"github.com/shortlink-org/shortlink/pkg/logger"
	"github.com/shortlink-org/shortlink/pkg/observability/metrics"

	// Import your auth service specific packages
	// permission_client "github.com/shortlink-org/auth/auth/internal/di/pkg/permission"
	// "github.com/shortlink-org/auth/auth/internal/services/permission"
)

// AuthServiceExample demonstrates how to structure auth service with default services
type AuthServiceExample struct {
	// Default services (automatically injected via DefaultSet)
	Log        logger.Logger
	Config     *config.Config
	AutoMaxPro autoMaxPro.AutoMaxPro
	Context    context.Context

	// Observability (automatically injected via DefaultSet)
	Tracer        trace.TracerProvider
	Metrics       *metrics.Monitoring
	PprofEndpoint profiling.PprofEndpoint

	// Auth-specific services
	authPermission *authzed.Client
	// permissionService *permission.Service
}

// AuthServiceSet provides all dependencies for AuthService
// This demonstrates two approaches:

// Approach 1: Use modular providers (recommended for new services)
var AuthServiceModularSet = wire.NewSet(
	// Default services
	providers.ProvideContext,
	providers.ProvideLogger,
	providers.ProvideConfig,
	providers.ProvideAutoMaxPro,
	providers.ProvideTracer,
	providers.ProvideMetrics,
	providers.ProvidePprofEndpoint,

	// Auth-specific providers
	ProvideAuthPermissionClient,
	// permission.New,

	// Service constructor
	NewAuthServiceExample,
)

// Approach 2: Use the consolidated DefaultSet (if you prefer)
var AuthServiceConsolidatedSet = wire.NewSet(
	DefaultSet, // This includes all default services

	// Auth-specific providers
	ProvideAuthPermissionClient,
	// permission.New,

	// Service constructor that accepts DefaultServices
	NewAuthServiceFromDefaults,
)

// Auth-specific provider functions
func ProvideAuthPermissionClient(config *config.Config) (*authzed.Client, error) {
	// Initialize authzed client with configuration
	// This is a placeholder - implement according to your needs
	return nil, nil
}

// NewAuthServiceExample creates auth service using individual dependencies
func NewAuthServiceExample(
	// Default services
	ctx context.Context,
	log logger.Logger,
	config *config.Config,
	autoMaxPro autoMaxPro.AutoMaxPro,

	// Observability
	tracer trace.TracerProvider,
	metrics *metrics.Monitoring,
	pprofEndpoint profiling.PprofEndpoint,

	// Auth-specific
	authPermission *authzed.Client,
	// permissionService *permission.Service,
) (*AuthServiceExample, error) {
	log.Info("Initializing AuthService")

	return &AuthServiceExample{
		Context:    ctx,
		Log:        log,
		Config:     config,
		AutoMaxPro: autoMaxPro,

		Tracer:        tracer,
		Metrics:       metrics,
		PprofEndpoint: pprofEndpoint,

		authPermission: authPermission,
		// permissionService: permissionService,
	}, nil
}

// NewAuthServiceFromDefaults creates auth service using DefaultServices struct
func NewAuthServiceFromDefaults(
	defaults *DefaultServices,
	authPermission *authzed.Client,
	// permissionService *permission.Service,
) (*AuthServiceExample, error) {
	defaults.Log.Info("Initializing AuthService from defaults")

	return &AuthServiceExample{
		Context:    defaults.Context,
		Log:        defaults.Log,
		Config:     defaults.Config,
		AutoMaxPro: defaults.AutoMaxPro,

		Tracer:        defaults.Tracer,
		Metrics:       defaults.Metrics,
		PprofEndpoint: defaults.PprofEndpoint,

		authPermission: authPermission,
		// permissionService: permissionService,
	}, nil
}

// Wire injector functions
func InitializeAuthServiceModular() (*AuthServiceExample, func(), error) {
	panic(wire.Build(AuthServiceModularSet))
}

func InitializeAuthServiceConsolidated() (*AuthServiceExample, func(), error) {
	panic(wire.Build(AuthServiceConsolidatedSet))
}