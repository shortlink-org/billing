# Default Services DI Package

This package provides a standardized dependency injection pattern for common services used across all microservices in the project.

## Overview

The default services package includes commonly used dependencies like:
- **Logger**: Structured logging
- **Config**: Application configuration
- **Context**: Request context
- **AutoMaxPro**: Automatic max process configuration
- **Tracer**: OpenTelemetry tracing
- **Metrics**: Monitoring and metrics collection
- **PprofEndpoint**: Performance profiling

## Structure

```
pkg/di/
├── default_services.go          # Main default services definition
├── providers/                   # Modular provider functions
│   ├── common.go               # Context and AutoMaxPro providers
│   ├── config.go               # Configuration provider
│   ├── logger.go               # Logger provider
│   └── observability.go        # Tracing, metrics, and profiling providers
├── example_service.go           # Example service implementation
├── auth_service_example.go      # Auth service integration example
└── README.md                   # This documentation
```

## Usage Patterns

### Pattern 1: Using Modular Providers (Recommended)

This approach gives you fine-grained control over which services to include:

```go
//go:generate go tool wire
//go:build wireinject

package your_service_di

import (
    "github.com/google/wire"
    "github.com/shortlink-org/shortlink/pkg/di/providers"
    // your service imports...
)

type YourService struct {
    // Common services
    Log        logger.Logger
    Config     *config.Config
    Context    context.Context
    
    // Observability
    Tracer     trace.TracerProvider
    Metrics    *metrics.Monitoring
    
    // Your service-specific dependencies
    Repository repository.Repository
    APIClient  client.APIClient
}

var YourServiceSet = wire.NewSet(
    // Default services (pick what you need)
    providers.ProvideContext,
    providers.ProvideLogger,
    providers.ProvideConfig,
    providers.ProvideTracer,
    providers.ProvideMetrics,
    
    // Your service-specific providers
    repository.New,
    client.New,
    
    // Service constructor
    NewYourService,
)

func NewYourService(
    ctx context.Context,
    log logger.Logger,
    config *config.Config,
    tracer trace.TracerProvider,
    metrics *metrics.Monitoring,
    repository repository.Repository,
    apiClient client.APIClient,
) (*YourService, error) {
    return &YourService{
        Context:    ctx,
        Log:        log,
        Config:     config,
        Tracer:     tracer,
        Metrics:    metrics,
        Repository: repository,
        APIClient:  apiClient,
    }, nil
}

func InitializeYourService() (*YourService, func(), error) {
    panic(wire.Build(YourServiceSet))
}
```

### Pattern 2: Using DefaultSet (For Complete Default Services)

This approach includes all default services at once:

```go
//go:generate go tool wire
//go:build wireinject

package your_service_di

import (
    "github.com/google/wire"
    "github.com/shortlink-org/shortlink/pkg/di"
    // your service imports...
)

type YourService struct {
    // Embed default services or access via composition
    *di.DefaultServices
    
    // Your service-specific dependencies
    Repository repository.Repository
    APIClient  client.APIClient
}

var YourServiceSet = wire.NewSet(
    di.DefaultSet, // Includes all default services
    
    // Your service-specific providers
    repository.New,
    client.New,
    
    // Service constructor
    NewYourService,
)

func NewYourService(
    defaults *di.DefaultServices,
    repository repository.Repository,
    apiClient client.APIClient,
) (*YourService, error) {
    defaults.Log.Info("Initializing YourService")
    
    return &YourService{
        DefaultServices: defaults,
        Repository:      repository,
        APIClient:       apiClient,
    }, nil
}

func InitializeYourService() (*YourService, func(), error) {
    panic(wire.Build(YourServiceSet))
}
```

## Migrating Existing Services

### Example: Updating Auth Service

Your original auth service:

```go
// Original auth service
var AuthSet = wire.NewSet(
    di.DefaultSet,
    permission_client.New,
    permission.New,
    NewAuthService,
)
```

Can be updated to use the new default services pattern:

```go
// Updated auth service with explicit providers
var AuthSet = wire.NewSet(
    // Default services
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
```

Or use the consolidated approach:

```go
// Updated auth service with DefaultSet
var AuthSet = wire.NewSet(
    di.DefaultSet, // All default services
    permission_client.New,
    permission.New,
    NewAuthService,
)
```

## Benefits

1. **Consistency**: All services use the same common dependencies
2. **Maintainability**: Changes to default services propagate automatically
3. **Modularity**: Pick only the services you need with the modular approach
4. **Testability**: Easy to mock default services for testing
5. **Documentation**: Clear separation between default and service-specific dependencies

## Adding New Default Services

To add a new default service:

1. Create a provider function in the appropriate file under `providers/`
2. Add the provider to the `DefaultSet` in `default_services.go`
3. Update the `DefaultServices` struct if you want it included there
4. Update documentation

Example:

```go
// In providers/new_service.go
func ProvideNewService() NewService {
    return NewService{}
}

// In default_services.go
var DefaultSet = wire.NewSet(
    // existing providers...
    providers.ProvideNewService,
    // ...
)
```

## Testing

When testing services that use default services, you can easily provide mocks:

```go
func TestYourService(t *testing.T) {
    // Create mocks for default services
    mockLogger := &mock.Logger{}
    mockConfig := &mock.Config{}
    
    // Create your service with mocks
    service := NewYourService(
        context.Background(),
        mockLogger,
        mockConfig,
        // ... other dependencies
    )
    
    // Test your service
}
```

## Wire Generation

Don't forget to generate the wire code after making changes:

```bash
cd pkg/di
go generate
```

This will create or update the `wire_gen.go` file with the actual implementation.