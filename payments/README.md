# Payments

<img width='200' height='200' src="./docs/public/logo.svg">

> [!NOTE]
> Service for work with payments.

### Getting started

We use Makefile for build and deploy.

```bash
$> make help # show help message with all commands and targets
```

### ADR

- [ADR-0001](./docs/ADR/decisions/0001-init.md) - Init project

### Use Cases

#### Payments

- [UC-1](./internal/usecases/create_payment/README.md) Create a payment for an invoice/order
- [UC-2](./internal/usecases/confirm_payment/README.md) Confirm a pending payment (SCA/3DS)
- [UC-3](./internal/usecases/capture_payment/README.md) Capture a previously authorized payment

#### Refunds

- [UC-4](./internal/usecases/refund_payment/README.md) Refund a payment (full or partial)

#### Subscriptions

- [UC-5](./internal/usecases/create_subscription/README.md) Create a new subscription for a customer
- [UC-6](./internal/usecases/cancel_subscription/README.md) Cancel an active subscription

#### Billing Cycle

- [UC-7](./internal/usecases/run_billing_cycle/README.md) Run scheduled subscription billing cycle

#### Webhooks

- [UC-8](./internal/usecases/handle_webhook/README.md) Handle provider webhook events idempotently
