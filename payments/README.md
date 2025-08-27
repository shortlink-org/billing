# Payments

<img width='200' height='200' src="./docs/public/logo.png">

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

- [UC-1](./internal/application/payments/usecase/create/README.md) Create a payment for an invoice/order
- [UC-2] _(Not implemented yet)_ Confirm a pending payment (SCA/3DS)
- [UC-3] _(Not implemented yet)_ Capture a previously authorized payment

#### Refunds

- [UC-4](./internal/application/payments/usecase/refund/README.md) Refund a payment (full or partial)

#### Subscriptions

- [UC-5] _(Not implemented yet)_ Create a new subscription for a customer
- [UC-6] _(Not implemented yet)_ Cancel an active subscription

#### Billing Cycle

- [UC-7] _(Not implemented yet)_ Run scheduled subscription billing cycle

#### Webhooks

- [UC-8] _(Not implemented yet)_ Handle provider webhook events idempotently
