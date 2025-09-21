# Stripe Payment Provider

This package provides integration with Stripe payment processing service.

## Configuration

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `STRIPE_API_KEY` | Stripe API key (starts with `sk_`) | Yes |

### Example Configuration

```bash
export STRIPE_API_KEY=sk_test_...
```

### API Key Types

- **Test keys**: Start with `sk_test_` - use for development and testing
- **Live keys**: Start with `sk_live_` - use for production

## Usage

The Stripe provider is automatically selected when:
- `PAYMENT_PROVIDER` is set to `stripe`
- `PAYMENT_PROVIDER` is not set (default behavior)

## Error Handling

- `ErrMissingAPIKey`: Returned when `STRIPE_API_KEY` environment variable is not set

## Implementation Details

The provider uses the official Stripe Go SDK (v82) and automatically configures the client with the provided API key.
