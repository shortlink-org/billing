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
- [UC-2](./#) Confirm a pending payment (SCA/3DS)
- [UC-3](./#) Capture a previously authorized payment

#### Refunds

- [UC-4](./internal/application/payments/usecase/refund/README.md) Refund a payment (full or partial)

#### Subscriptions

- [UC-5](./#) Create a new subscription for a customer
- [UC-6](./#) Cancel an active subscription

#### Billing Cycle

- [UC-7](./#) Run scheduled subscription billing cycle

#### Webhooks

- [UC-8](./#) Handle provider webhook events idempotently
