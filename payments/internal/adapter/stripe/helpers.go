package stripeadp

import (
	"github.com/stripe/stripe-go/v82"
	"github.com/shortlink-org/billing/payments/internal/domain/payment/ledger"
	"google.golang.org/genproto/googleapis/type/money"
)

// fromMinor converts minor currency units to money.Money.
func fromMinor(cur stripe.Currency, v int64) *money.Money {
	return ledger.MinorUnitsToAmount(string(cur), v)
}