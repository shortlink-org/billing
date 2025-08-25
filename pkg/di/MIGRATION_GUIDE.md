# Migration Guide: Updating Services to Use Default Services Pattern

This guide shows how to update your existing services to use the new default services pattern.

## Your Current Auth Service

```go
//go:generate go tool wire
//go:build wireinject

package auth_di

import (
	"github.com/authzed/authzed-go/v1"
	"github.com/google/wire"
	"go.opentelemetry.io/otel/trace"

	"github.com/shortlink-org/shortlink/pkg/di"
	"github.com/shortlink-org/shortlink/pkg/di/pkg/autoMaxPro"
	"github.com/shortlink-org/shortlink/pkg/di/pkg/config"
	"github.com/shortlink-org/shortlink/pkg/di/pkg/profiling"
	"github.com/shortlink-org/shortlink/pkg/logger"
	"github.com/shortlink-org/shortlink/pkg/observability/metrics"

	permission_client "github.com/shortlink-org/auth/auth/internal/di/pkg/permission"
	"github.com/shortlink-org/auth/auth/internal/services/permission"
)

type AuthService struct {
	// Common
	Log        logger.Logger
	Config     *config.Config
	AutoMaxPro autoMaxPro.AutoMaxPro

	// Observability
	Tracer        trace.TracerProvider
	Metrics       *metrics.Monitoring
	PprofEndpoint profiling.PprofEndpoint

	// Security
	authPermission *authzed.Client

	// Application
	permissionService *permission.Service
}

var AuthSet = wire.NewSet(
	di.DefaultSet,
	permission_client.New,
	permission.New,
	NewAuthService,
)

func NewAuthService(
	log logger.Logger,
	config *config.Config,
	autoMaxProcsOption autoMaxPro.AutoMaxPro,
	metrics *metrics.Monitoring,
	tracer trace.TracerProvider,
	pprofHTTP profiling.PprofEndpoint,
	authPermission *authzed.Client,
	permissionService *permission.Service,
) (*AuthService, error) {
	return &AuthService{
		Log:    log,
		Config: config,

		Tracer:        tracer,
		Metrics:       metrics,
		PprofEndpoint: pprofHTTP,
		AutoMaxPro:    autoMaxProcsOption,

		authPermission: authPermission,

		permissionService: permissionService,
	}, nil
}

func InitializeAuthService() (*AuthService, func(), error) {
	panic(wire.Build(AuthSet))
}
```

## Migration Option 1: Use Modular Providers (Recommended)

This approach gives you explicit control over which default services to include:

```go
//go:generate go tool wire
//go:build wireinject

package auth_di

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

	permission_client "github.com/shortlink-org/auth/auth/internal/di/pkg/permission"
	"github.com/shortlink-org/auth/auth/internal/services/permission"
)

type AuthService struct {
	// Common services
	Log        logger.Logger
	Config     *config.Config
	AutoMaxPro autoMaxPro.AutoMaxPro
	Context    context.Context

	// Observability
	Tracer        trace.TracerProvider
	Metrics       *metrics.Monitoring
	PprofEndpoint profiling.PprofEndpoint

	// Security
	authPermission *authzed.Client

	// Application
	permissionService *permission.Service
}

// Updated AuthSet using modular providers
var AuthSet = wire.NewSet(
	// Default services (explicit)
	providers.ProvideContext,
	providers.ProvideLogger,
	providers.ProvideConfig,
	providers.ProvideAutoMaxPro,
	providers.ProvideTracer,
	providers.ProvideMetrics,
	providers.ProvidePprofEndpoint,

	// Auth-specific services
	permission_client.New,
	permission.New,

	NewAuthService,
)

// Updated constructor with context added
func NewAuthService(
	// Common services
	ctx context.Context,
	log logger.Logger,
	config *config.Config,
	autoMaxProcsOption autoMaxPro.AutoMaxPro,

	// Observability
	tracer trace.TracerProvider,
	metrics *metrics.Monitoring,
	pprofHTTP profiling.PprofEndpoint,

	// Auth-specific
	authPermission *authzed.Client,
	permissionService *permission.Service,
) (*AuthService, error) {
	log.Info("Initializing AuthService")

	return &AuthService{
		// Common services
		Context:    ctx,
		Log:        log,
		Config:     config,
		AutoMaxPro: autoMaxProcsOption,

		// Observability
		Tracer:        tracer,
		Metrics:       metrics,
		PprofEndpoint: pprofHTTP,

		// Auth-specific
		authPermission:    authPermission,
		permissionService: permissionService,
	}, nil
}

func InitializeAuthService() (*AuthService, func(), error) {
	panic(wire.Build(AuthSet))
}
```

## Migration Option 2: Use DefaultSet (Minimal Changes)

This approach keeps your existing structure but uses the new DefaultSet:

```go
//go:generate go tool wire
//go:build wireinject

package auth_di

import (
	"github.com/authzed/authzed-go/v1"
	"github.com/google/wire"
	"go.opentelemetry.io/otel/trace"

	"github.com/shortlink-org/shortlink/pkg/di"
	"github.com/shortlink-org/shortlink/pkg/di/pkg/autoMaxPro"
	"github.com/shortlink-org/shortlink/pkg/di/pkg/config"
	"github.com/shortlink-org/shortlink/pkg/di/pkg/profiling"
	"github.com/shortlink-org/shortlink/pkg/logger"
	"github.com/shortlink-org/shortlink/pkg/observability/metrics"

	permission_client "github.com/shortlink-org/auth/auth/internal/di/pkg/permission"
	"github.com/shortlink-org/auth/auth/internal/services/permission"
)

type AuthService struct {
	// Common
	Log        logger.Logger
	Config     *config.Config
	AutoMaxPro autoMaxPro.AutoMaxPro

	// Observability
	Tracer        trace.TracerProvider
	Metrics       *metrics.Monitoring
	PprofEndpoint profiling.PprofEndpoint

	// Security
	authPermission *authzed.Client

	// Application
	permissionService *permission.Service
}

// Minimal change: just update the import to use the new DefaultSet
var AuthSet = wire.NewSet(
	di.DefaultSet, // This now uses the new default services package
	permission_client.New,
	permission.New,
	NewAuthService,
)

// Keep your existing constructor unchanged
func NewAuthService(
	log logger.Logger,
	config *config.Config,
	autoMaxProcsOption autoMaxPro.AutoMaxPro,
	metrics *metrics.Monitoring,
	tracer trace.TracerProvider,
	pprofHTTP profiling.PprofEndpoint,
	authPermission *authzed.Client,
	permissionService *permission.Service,
) (*AuthService, error) {
	return &AuthService{
		Log:        log,
		Config:     config,
		AutoMaxPro: autoMaxProcsOption,

		Tracer:        tracer,
		Metrics:       metrics,
		PprofEndpoint: pprofHTTP,

		authPermission:    authPermission,
		permissionService: permissionService,
	}, nil
}

func InitializeAuthService() (*AuthService, func(), error) {
	panic(wire.Build(AuthSet))
}
```

## Migration Option 3: Use DefaultServices Struct

This approach uses the DefaultServices struct for cleaner dependency management:

```go
//go:generate go tool wire
//go:build wireinject

package auth_di

import (
	"github.com/authzed/authzed-go/v1"
	"github.com/google/wire"

	"github.com/shortlink-org/shortlink/pkg/di"
	permission_client "github.com/shortlink-org/auth/auth/internal/di/pkg/permission"
	"github.com/shortlink-org/auth/auth/internal/services/permission"
)

type AuthService struct {
	// Embed default services
	*di.DefaultServices

	// Auth-specific services
	authPermission    *authzed.Client
	permissionService *permission.Service
}

var AuthSet = wire.NewSet(
	di.DefaultSet,
	permission_client.New,
	permission.New,
	NewAuthService,
)

// Simplified constructor using DefaultServices
func NewAuthService(
	defaults *di.DefaultServices,
	authPermission *authzed.Client,
	permissionService *permission.Service,
) (*AuthService, error) {
	defaults.Log.Info("Initializing AuthService")

	return &AuthService{
		DefaultServices:   defaults,
		authPermission:    authPermission,
		permissionService: permissionService,
	}, nil
}

func InitializeAuthService() (*AuthService, func(), error) {
	panic(wire.Build(AuthSet))
}
```

## Which Option to Choose?

### Option 1 (Modular Providers) - **Recommended for new services**
- ✅ Explicit control over dependencies
- ✅ Easy to see what services are included
- ✅ Better for testing (can mock individual services)
- ✅ Future-proof
- ❌ More verbose

### Option 2 (Minimal Changes) - **Good for existing services**
- ✅ Minimal code changes required
- ✅ Maintains existing patterns
- ✅ Easy migration path
- ❌ Less explicit about dependencies
- ❌ Includes all default services whether needed or not

### Option 3 (DefaultServices Struct) - **Good for services needing all defaults**
- ✅ Very clean and simple
- ✅ Easy access to all default services
- ✅ Reduces constructor complexity
- ❌ Tight coupling to DefaultServices struct
- ❌ Less flexibility

## Migration Steps

1. **Update imports**: Add the new default services package import
2. **Choose your approach**: Pick one of the three options above
3. **Update your wire set**: Replace `di.DefaultSet` with your chosen pattern
4. **Update constructor**: Modify function signature if needed
5. **Update struct**: Add any new fields (like Context) if using Option 1
6. **Generate wire code**: Run `go generate` to update wire_gen.go
7. **Test**: Ensure everything compiles and works correctly

## Testing Your Migration

After migration, test that:
1. The service still compiles
2. Wire generation works: `go generate`
3. All dependencies are properly injected
4. The service starts up correctly
5. All functionality works as expected

## Common Issues

1. **Missing context import**: If using Option 1, don't forget to import `context`
2. **Constructor parameter order**: Make sure parameters match the wire providers
3. **Wire generation**: Always run `go generate` after changes
4. **Unused imports**: Clean up any imports that are no longer needed