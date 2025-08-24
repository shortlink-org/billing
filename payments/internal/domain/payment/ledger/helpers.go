package ledger

import (
	"google.golang.org/genproto/googleapis/type/money"
)

const nanosInUnit int64 = 1_000_000_000

// ensureMoney returns m or a zero-money in the requested currency.
func ensureMoney(currency string, m *money.Money) *money.Money {
	if m != nil {
		return Clone(m)
	}
	return &money.Money{CurrencyCode: currency, Units: 0, Nanos: 0}
}

func validatePositive(m *money.Money) error {
	if m == nil {
		return ErrNilMoney
	}
	// forbid zero for operations
	if m.GetUnits() == 0 && m.GetNanos() == 0 {
		return ErrNonPositiveAmount
	}
	// forbid negative
	if m.GetUnits() < 0 || (m.GetUnits() == 0 && m.GetNanos() < 0) {
		return ErrNonPositiveAmount
	}
	return nil
}

func sameCurrency(a, b *money.Money) error {
	if a.GetCurrencyCode() != b.GetCurrencyCode() {
		return ErrCurrencyMismatch
	}
	return nil
}

func Clone(m *money.Money) *money.Money {
	if m == nil {
		return nil
	}
	c := *m
	return &c
}

// add = a + b (same currency)
func add(a, b *money.Money) *money.Money {
	n := toNanos(a) + toNanos(b)
	return fromNanos(a.GetCurrencyCode(), n)
}

// cmp: -1 if a<b, 0 if ==, 1 if a>b
func cmp(a, b *money.Money) int {
	na := toNanos(a)
	nb := toNanos(b)
	switch {
	case na < nb:
		return -1
	case na > nb:
		return 1
	default:
		return 0
	}
}

func toNanos(m *money.Money) int64 {
	return m.GetUnits()*nanosInUnit + int64(m.GetNanos())
}

func fromNanos(currency string, nanos int64) *money.Money {
	units := nanos / nanosInUnit
	rem := nanos % nanosInUnit
	// normalize so sign of nanos matches units and |nanos| < 1e9
	if rem < 0 && units > 0 {
		units--
		rem += nanosInUnit
	}
	if rem > 0 && units < 0 {
		units++
		rem -= nanosInUnit
	}
	return &money.Money{
		CurrencyCode: currency,
		Units:        units,
		Nanos:        int32(rem),
	}
}
