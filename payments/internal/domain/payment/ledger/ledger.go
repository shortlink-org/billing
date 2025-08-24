package ledger

import (
	"google.golang.org/genproto/googleapis/type/money"
)

// Ledger is a Value Object holding monetary totals for a payment.
// All amounts must share the same currency (as Amount).
type Ledger struct {
	Amount        *money.Money // charge target
	Authorized    *money.Money // total hold
	Captured      *money.Money // total captured
	TotalRefunded *money.Money // total refunded
}

// Authorize accumulates a hold.
// Invariants: amt > 0, same currency, Authorized+amt <= Amount.
func (l *Ledger) Authorize(amt *money.Money) error {
	if l.Amount == nil {
		return ErrNilAmount
	}
	if err := validatePositive(amt); err != nil {
		return err
	}
	if err := sameCurrency(l.Amount, amt); err != nil {
		return err
	}

	curAuth := ensureMoney(l.Amount.GetCurrencyCode(), l.Authorized)
	next := add(curAuth, amt)
	if cmp(next, l.Amount) > 0 {
		return ErrAuthorizeExceeds
	}

	l.Authorized = next
	return nil
}

// Capture accumulates captured total.
// Invariants: amt > 0, same currency.
// Limit: if Authorized present -> Captured+amt <= Authorized; else <= Amount.
func (l *Ledger) Capture(amt *money.Money) error {
	if l.Amount == nil {
		return ErrNilAmount
	}
	if err := validatePositive(amt); err != nil {
		return err
	}
	if err := sameCurrency(l.Amount, amt); err != nil {
		return err
	}

	curCap := ensureMoney(l.Amount.GetCurrencyCode(), l.Captured)
	next := add(curCap, amt)

	limit := l.Amount
	if l.Authorized != nil {
		limit = l.Authorized
	}
	if cmp(next, limit) > 0 {
		return ErrCaptureExceedsLimit
	}

	l.Captured = next
	return nil
}

// Refund accumulates TotalRefunded.
// Invariants: amt > 0, same currency, TotalRefunded+amt <= Captured.
// Returns full=true if after the operation TotalRefunded == Captured.
func (l *Ledger) Refund(amt *money.Money) (bool, error) {
	if l.Captured == nil {
		return false, ErrRefundWithoutCapture
	}
	if err := validatePositive(amt); err != nil {
		return false, err
	}
	if err := sameCurrency(l.Captured, amt); err != nil {
		return false, err
	}

	curRef := ensureMoney(l.Captured.GetCurrencyCode(), l.TotalRefunded)
	next := add(curRef, amt)
	if cmp(next, l.Captured) > 0 {
		return false, ErrRefundExceeds
	}

	l.TotalRefunded = next
	return cmp(l.TotalRefunded, l.Captured) == 0, nil
}
