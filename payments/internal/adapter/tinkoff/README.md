# Tinkoff Payment Provider

This package provides integration with Tinkoff Bank payment processing service.

## Configuration

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `TINKOFF_API_KEY` | Tinkoff API key for authentication | Yes |
| `TINKOFF_CERT_FILE` | Path to client certificate file (.pem) | Yes |
| `TINKOFF_KEY_FILE` | Path to client private key file (.key) | Yes |
| `TINKOFF_BASE_URL` | Base URL for Tinkoff API | No (defaults to https://secured-openapi.tbank.ru) |

### Example Configuration

```bash
export TINKOFF_API_KEY=your_tinkoff_api_key
export TINKOFF_CERT_FILE=/path/to/your/cert.pem
export TINKOFF_KEY_FILE=/path/to/your/key.key
export TINKOFF_BASE_URL=https://secured-openapi.tbank.ru
```

## Setup

To use Tinkoff as a payment provider, you need to:

1. Obtain client certificates from Tinkoff
2. Set up the required environment variables
3. Configure the payment details (currently hardcoded for demo purposes)

## Usage

The Tinkoff provider is selected when `PAYMENT_PROVIDER` is set to `tinkoff`:

```bash
export PAYMENT_PROVIDER=tinkoff
```

## Error Handling

- `ErrMissingAPIKey`: Returned when `TINKOFF_API_KEY` environment variable is not set
- `ErrMissingCertFile`: Returned when `TINKOFF_CERT_FILE` environment variable is not set
- `ErrMissingKeyFile`: Returned when `TINKOFF_KEY_FILE` environment variable is not set

## Production Configuration

The following fields in the Tinkoff payment request are currently hardcoded and should be configured for production:

- `From.AccountNumber` - Your Tinkoff account number
- `To.Name` - Recipient name
- `To.INN` - Recipient INN
- `To.KPP` - Recipient KPP
- `To.BIK` - Bank BIK
- `To.BankName` - Bank name
- `To.CorrAccountNumber` - Correspondent account number
- `To.AccountNumber` - Recipient account number
- `Tax.KBK` - Budget classification code
- `Tax.OKTMO` - OKTMO code

These should be moved to configuration files or environment variables for production use.

## Implementation Details

The provider uses mutual TLS authentication with client certificates and makes HTTP requests to the Tinkoff API endpoints.
