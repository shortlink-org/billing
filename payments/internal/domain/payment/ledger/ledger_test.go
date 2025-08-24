package ledger

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/type/money"
)

func M(cur string, units int64, nanos int32) *money.Money {
	return &money.Money{CurrencyCode: cur, Units: units, Nanos: nanos}
}

func TestAuthorizeAccumulatesAndCapsAtAmount(t *testing.T) {
	l := &Ledger{Amount: M("USD", 10, 0)} // $10.00

	require.NoError(t, l.Authorize(M("USD", 3, 0)))
	require.Equal(t, int64(3), l.Authorized.Units)

	require.NoError(t, l.Authorize(M("USD", 7, 0)))
	require.Equal(t, int64(10), l.Authorized.Units)

	err := l.Authorize(M("USD", 1, 0))
	require.ErrorIs(t, err, ErrAuthorizeExceeds)
}

func TestAuthorizeValidations(t *testing.T) {
	l := &Ledger{Amount: M("USD", 5, 0)}

	require.ErrorIs(t, l.Authorize(nil), ErrNilMoney)
	require.ErrorIs(t, l.Authorize(M("USD", 0, 0)), ErrNonPositiveAmount)
	require.ErrorIs(t, l.Authorize(M("USD", -1, 0)), ErrNonPositiveAmount)
	require.ErrorIs(t, l.Authorize(M("EUR", 1, 0)), ErrCurrencyMismatch)

	// amount not set
	l2 := &Ledger{}
	require.ErrorIs(t, l2.Authorize(M("USD", 1, 0)), ErrNilAmount)
}

func TestCaptureWithAuthorized(t *testing.T) {
	l := &Ledger{Amount: M("USD", 10, 0)}
	require.NoError(t, l.Authorize(M("USD", 10, 0)))

	require.NoError(t, l.Capture(M("USD", 4, 0)))
	require.Equal(t, int64(4), l.Captured.Units)

	require.NoError(t, l.Capture(M("USD", 6, 0)))
	require.Equal(t, int64(10), l.Captured.Units)

	err := l.Capture(M("USD", 1, 0))
	require.ErrorIs(t, err, ErrCaptureExceedsLimit)
}

func TestCaptureImmediateWithoutAuthorized(t *testing.T) {
	l := &Ledger{Amount: M("USD", 10, 0)}

	require.NoError(t, l.Capture(M("USD", 3, 0)))
	require.Equal(t, int64(3), l.Captured.Units)

	require.NoError(t, l.Capture(M("USD", 7, 0)))
	require.Equal(t, int64(10), l.Captured.Units)

	err := l.Capture(M("USD", 1, 0))
	require.ErrorIs(t, err, ErrCaptureExceedsLimit)
}

func TestRefundAccumulatesAndStopsAtCaptured(t *testing.T) {
	l := &Ledger{
		Amount:   M("USD", 10, 0),
		Captured: M("USD", 10, 0),
	}

	full, err := l.Refund(M("USD", 3, 0))
	require.NoError(t, err)
	require.False(t, full)
	require.Equal(t, int64(3), l.TotalRefunded.Units)

	full, err = l.Refund(M("USD", 7, 0))
	require.NoError(t, err)
	require.True(t, full)
	require.Equal(t, int64(10), l.TotalRefunded.Units)

	_, err = l.Refund(M("USD", 1, 0))
	require.ErrorIs(t, err, ErrRefundExceeds)
}

// ---- Scale validation -------------------------------------------------------

func TestScaleUSD(t *testing.T) {
	l := &Ledger{Amount: M("USD", 1, 0)} // USD has 2 decimals

	require.ErrorIs(t, l.Authorize(M("USD", 0, 5_000_000)), ErrInvalidScale) // 0.005 not allowed
	require.NoError(t, l.Authorize(M("USD", 0, 10_000_000)))                 // 0.01 ok
}

func TestScaleJPY(t *testing.T) {
	l := &Ledger{Amount: M("JPY", 100, 0)} // JPY has 0 decimals

	require.ErrorIs(t, l.Authorize(M("JPY", 0, 1)), ErrInvalidScale) // nanos must be 0
	require.NoError(t, l.Authorize(M("JPY", 10, 0)))
}

func TestScaleKWD(t *testing.T) {
	l := &Ledger{Amount: M("KWD", 1, 0)} // KWD has 3 decimals (step=1e6 nanos)

	require.NoError(t, l.Authorize(M("KWD", 0, 5_000_000)))                // 0.005 ok
	require.ErrorIs(t, l.Authorize(M("KWD", 0, 500_000)), ErrInvalidScale) // 0.0005 not ok
}

// ---- Queries ---------------------------------------------------------------

func TestRemainingToCaptureAndRefundable(t *testing.T) {
	l := &Ledger{
		Amount:     M("USD", 10, 0),
		Authorized: M("USD", 8, 0),
		Captured:   M("USD", 3, 0),
	}

	rem := l.RemainingToCapture() // min(Amount, Authorized) - Captured = 8 - 3 = 5
	require.Equal(t, int64(5), rem.Units)
	require.Equal(t, int32(0), rem.Nanos)

	refundable := l.Refundable() // Captured - TotalRefunded (nil -> 0) = 3
	require.Equal(t, int64(3), refundable.Units)
	require.Equal(t, int32(0), refundable.Nanos)

	// Refund part, check IsFullyRefunded and Refundable
	full, err := l.Refund(M("USD", 2, 0))
	require.NoError(t, err)
	require.False(t, full)
	require.False(t, l.IsFullyRefunded())

	refundable = l.Refundable() // 3 - 2 = 1
	require.Equal(t, int64(1), refundable.Units)

	full, err = l.Refund(M("USD", 1, 0))
	require.NoError(t, err)
	require.True(t, full)
	require.True(t, l.IsFullyRefunded())
}

// ---- Nanos arithmetic (valid scale) ----------------------------------------

func TestSubUnitsWithNanos_ValidScale(t *testing.T) {
	l := &Ledger{Amount: M("USD", 0, 90_000_000)} // cap = $0.09 (valid for USD)

	require.NoError(t, l.Authorize(M("USD", 0, 50_000_000))) // $0.05
	require.NoError(t, l.Authorize(M("USD", 0, 40_000_000))) // +$0.04
	require.Equal(t, int32(90_000_000), l.Authorized.Nanos)  // $0.09

	// попытка превысить cap на 0.01 — должна упасть
	err := l.Authorize(M("USD", 0, 10_000_000))
	require.ErrorIs(t, err, ErrAuthorizeExceeds)
}
