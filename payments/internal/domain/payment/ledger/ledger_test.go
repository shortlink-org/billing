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

func TestRefundValidations(t *testing.T) {
	l := &Ledger{
		Amount: M("USD", 10, 0),
	}
	_, err := l.Refund(M("USD", 1, 0))
	require.ErrorIs(t, err, ErrRefundWithoutCapture)

	l.Captured = M("USD", 5, 0)

	_, err = l.Refund(nil)
	require.ErrorIs(t, err, ErrNilMoney)

	_, err = l.Refund(M("USD", 0, 0))
	require.ErrorIs(t, err, ErrNonPositiveAmount)

	_, err = l.Refund(M("EUR", 1, 0))
	require.ErrorIs(t, err, ErrCurrencyMismatch)
}

func TestSubUnitsWithNanos(t *testing.T) {
	l := &Ledger{Amount: M("USD", 0, 95_000_000)} // cap = $0.095

	require.NoError(t, l.Authorize(M("USD", 0, 50_000_000))) // $0.05
	require.NoError(t, l.Authorize(M("USD", 0, 45_000_000))) // +$0.045
	require.Equal(t, int32(95_000_000), l.Authorized.Nanos)  // $0.095

	// попытка превысить cap на 1 нанос — должна упасть
	err := l.Authorize(M("USD", 0, 1))
	require.ErrorIs(t, err, ErrAuthorizeExceeds)
}
