package ledger

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	"google.golang.org/genproto/googleapis/type/money"
)

// ==== currency scale (minor units) ===========================================

// Known ISO-4217 exponents (extend as needed).
var currencyExp = map[string]int{
	"USD": 2, "EUR": 2, "RUB": 2, "GBP": 2, "CHF": 2,
	"JPY": 0,
	"KWD": 3, "BHD": 3, "JOD": 3,
}

// RegisterCurrencyExponent allows adding/updating currency exponent at runtime.
// exp must be in [0,9].
func RegisterCurrencyExponent(code string, exp int) error {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return fmt.Errorf("money: empty currency code")
	}
	if exp < 0 || exp > 9 {
		return fmt.Errorf("money: invalid exponent %d for %s", exp, code)
	}
	currencyExp[code] = exp
	return nil
}

// ScaleOf returns ISO-4217 exponent (minor units).
func ScaleOf(code string) (int, error) {
	exp, ok := currencyExp[strings.ToUpper(strings.TrimSpace(code))]
	if !ok {
		return 0, ErrUnknownCurrency
	}
	return exp, nil
}

// StepNanos returns allowed nanos step (10^(9-exp)).
func StepNanos(code string) (int32, error) {
	exp, err := ScaleOf(code)
	if err != nil {
		return 0, err
	}
	step := int32(1)
	for i := 0; i < 9-exp; i++ {
		step *= 10
	}
	return step, nil
}

// ==== decimal <-> money ======================================================

var oneE9 = decimal.NewFromInt(1_000_000_000)

// normalizeUnitsNanos converts (units, nanos) to single signed nanounits.
func normalizeUnitsNanos(units int64, nanos int32) int64 {
	total := units*1_000_000_000 + int64(nanos)
	return total
}

// moneyToScaled returns amount as integer decimal of nanounits.
func moneyToScaled(m *money.Money) (decimal.Decimal, error) {
	if m == nil {
		return decimal.Zero, ErrNilMoney
	}
	if err := validateScale(m); err != nil {
		return decimal.Zero, err
	}
	n := normalizeUnitsNanos(m.GetUnits(), m.GetNanos())
	return decimal.NewFromInt(n), nil
}

// scaledToMoney converts integer nanounits to google.type.Money (no rounding!).
func scaledToMoney(currency string, scaled decimal.Decimal) (*money.Money, error) {
	if !scaled.IsInteger() {
		// we always keep integer nanounits
		return nil, fmt.Errorf("money: scaled value must be integer nanounits")
	}
	nanosInt := scaled.IntPart() // nanounits
	units := nanosInt / 1_000_000_000
	rem := nanosInt % 1_000_000_000
	// keep sign normalized (google.type.Money allows mixed signs but we normalize)
	if rem < 0 && units > 0 {
		units--
		rem += 1_000_000_000
	}
	if rem > 0 && units < 0 {
		units++
		rem -= 1_000_000_000
	}
	res := &money.Money{
		CurrencyCode: strings.ToUpper(currency),
		Units:        units,
		Nanos:        int32(rem),
	}
	// final scale check (must match ISO exponent)
	if err := validateScale(res); err != nil {
		return nil, err
	}
	return res, nil
}

// ==== validations & small helpers ===========================================

func Zero(code string) *money.Money {
	return &money.Money{CurrencyCode: strings.ToUpper(code)}
}

func Clone(m *money.Money) *money.Money {
	if m == nil {
		return nil
	}
	cp := *m
	return &cp
}

func Currency(m *money.Money) string { return m.GetCurrencyCode() }

func validateScale(m *money.Money) error {
	cur := m.GetCurrencyCode()
	step, err := StepNanos(cur)
	if err != nil {
		return err
	}
	if m.GetNanos()%step != 0 {
		return ErrInvalidScale
	}
	return nil
}

func sameCurrency(a, b *money.Money) error {
	if a.GetCurrencyCode() != b.GetCurrencyCode() {
		return ErrCurrencyMismatch
	}
	return nil
}

func validatePositive(m *money.Money) error {
	if m == nil {
		return ErrNilMoney
	}
	if m.GetUnits() == 0 && m.GetNanos() == 0 {
		return ErrNonPositiveAmount
	}
	// negative?
	if m.GetUnits() < 0 || (m.GetUnits() == 0 && m.GetNanos() < 0) {
		return ErrNonPositiveAmount
	}
	return validateScale(m)
}

func validateAuthorizeInput(amount, amt *money.Money) error {
	if amount == nil {
		return ErrNilAmount
	}
	if err := validatePositive(amt); err != nil {
		return err
	}
	return sameCurrency(amount, amt)
}

func ensureMoney(currency string, m *money.Money) *money.Money {
	if m != nil {
		return Clone(m)
	}
	return Zero(currency)
}

// ==== arithmetic API (public, with decimal under the hood) ===================

// Compare returns -1 if a<b, 0 if equal, 1 if a>b (same currency required).
func Compare(a, b *money.Money) int {
	an, _ := moneyToScaled(a)
	bn, _ := moneyToScaled(b)
	return an.Cmp(bn)
}

// Add returns a+b (same currency & valid scale). No rounding happens.
func Add(a, b *money.Money) (*money.Money, error) {
	if a == nil || b == nil {
		return nil, ErrNilMoney
	}
	if err := sameCurrency(a, b); err != nil {
		return nil, err
	}
	if err := validateScale(a); err != nil {
		return nil, err
	}
	if err := validateScale(b); err != nil {
		return nil, err
	}
	an, _ := moneyToScaled(a)
	bn, _ := moneyToScaled(b)
	sum := an.Add(bn)
	return scaledToMoney(a.GetCurrencyCode(), sum)
}

// Sub returns a-b (same currency & valid scale). Result may be negative.
func Sub(a, b *money.Money) (*money.Money, error) {
	if a == nil || b == nil {
		return nil, ErrNilMoney
	}
	if err := sameCurrency(a, b); err != nil {
		return nil, err
	}
	if err := validateScale(a); err != nil {
		return nil, err
	}
	if err := validateScale(b); err != nil {
		return nil, err
	}
	an, _ := moneyToScaled(a)
	bn, _ := moneyToScaled(b)
	diff := an.Sub(bn)
	return scaledToMoney(a.GetCurrencyCode(), diff)
}

// AmountToMinorUnits converts Money to integer minor units (e.g., cents).
// It validates currency scale and guarantees exact conversion (no rounding).
func AmountToMinorUnits(m *money.Money) (int64, error) {
	if m == nil {
		return 0, ErrNilMoney
	}
	if err := validateScale(m); err != nil {
		return 0, err
	}
	exp, err := ScaleOf(m.GetCurrencyCode())
	if err != nil {
		return 0, err
	}
	nanos, err := moneyToScaled(m) // integer nanounits (1e-9)
	if err != nil {
		return 0, err
	}
	// minor = nanos / 10^(9-exp)
	div := decimal.NewFromInt(1).Shift(int32(9 - exp))
	minor := nanos.Div(div)
	if !minor.IsInteger() {
		return 0, ErrInvalidScale // should not happen due to validateScale
	}
	return minor.IntPart(), nil
}

// MinorUnitsToAmount converts integer minor units (e.g., cents) to Money.
// The result respects the ISO-4217 scale (nanos step = 10^(9-exp)).
func MinorUnitsToAmount(currency string, v int64) *money.Money {
	exp, _ := ScaleOf(currency)
	// nanos = minor * 10^(9-exp)
	nanos := decimal.NewFromInt(v).Mul(decimal.NewFromInt(1).Shift(int32(9 - exp)))
	res, _ := scaledToMoney(currency, nanos) // nanos is integer → no error
	return res
}
