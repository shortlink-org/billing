package ledger

import (
	"google.golang.org/genproto/googleapis/type/money"
)

// Ledger is a Value Object holding monetary totals for a payment.
// All amounts must share the same currency (as Amount).
type Ledger struct {
	Amount        *money.Money // target to charge
	Authorized    *money.Money // total hold
	Captured      *money.Money // total captured
	TotalRefunded *money.Money // total refunded
}

// Authorize accumulates a hold.
// Invariants: amt > 0, same currency, Authorized+amt <= Amount.
func (l *Ledger) Authorize(amt *money.Money) error {
	if err := validateAuthorizeInput(l.Amount, amt); err != nil {
		return err
	}
	var next *money.Money
	if l.Authorized == nil {
		next = Clone(amt)
	} else {
		s, err := Add(l.Authorized, amt)
		if err != nil {
			return err
		}
		next = s
	}
	// Hard cap: Authorized <= Amount
	if Compare(next, l.Amount) > 0 {
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
	next, err := Add(curCap, amt)
	if err != nil {
		return err
	}

	limit := l.Amount
	if l.Authorized != nil {
		limit = l.Authorized
	}
	if Compare(next, limit) > 0 {
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
	next, err := Add(curRef, amt)
	if err != nil {
		return false, err
	}
	if Compare(next, l.Captured) > 0 {
		return false, ErrRefundExceeds
	}

	l.TotalRefunded = next
	return Compare(l.TotalRefunded, l.Captured) == 0, nil
}

// RemainingToCapture returns Amount/Authorized minus Captured (same currency).
func (l *Ledger) RemainingToCapture() *money.Money {
	if l.Amount == nil {
		return nil
	}
	limit := l.Amount
	if l.Authorized != nil {
		limit = l.Authorized
	}
	cap := ensureMoney(limit.GetCurrencyCode(), l.Captured)
	diff, _ := Sub(limit, cap)
	return diff
}

// Refundable returns Captured - TotalRefunded (same currency).
func (l *Ledger) Refundable() *money.Money {
	if l.Captured == nil {
		return nil
	}
	ref := ensureMoney(l.Captured.GetCurrencyCode(), l.TotalRefunded)
	diff, _ := Sub(l.Captured, ref)
	return diff
}

// IsFullyRefunded is true if TotalRefunded == Captured (>0).
func (l *Ledger) IsFullyRefunded() bool {
	if l.Captured == nil {
		return false
	}
	return Compare(ensureMoney(l.Captured.GetCurrencyCode(), l.TotalRefunded), l.Captured) == 0
}
