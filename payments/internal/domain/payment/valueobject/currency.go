package valueobject

import (
	"errors"
	"strings"
)

var (
	// ErrInvalidCurrencyCode is returned when currency code is invalid.
	ErrInvalidCurrencyCode = errors.New("currency: invalid currency code")
	// ErrUnsupportedCurrency is returned when currency is not supported.
	ErrUnsupportedCurrency = errors.New("currency: unsupported currency")
)

// Currency represents an ISO 4217 currency code as a value object.
type Currency struct {
	code string
}

// NewCurrency creates a new Currency value object with validation.
func NewCurrency(code string) (Currency, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if len(code) != 3 {
		return Currency{}, ErrInvalidCurrencyCode
	}

	// Validate against known currencies (extend as needed)
	if !isValidCurrencyCode(code) {
		return Currency{}, ErrUnsupportedCurrency
	}

	return Currency{code: code}, nil
}

// MustNewCurrency creates a new Currency value object, panicking on error.
// Use only when you're certain the currency code is valid.
func MustNewCurrency(code string) Currency {
	c, err := NewCurrency(code)
	if err != nil {
		panic(err)
	}
	return c
}

// Code returns the currency code.
func (c Currency) Code() string {
	return c.code
}

// String returns the string representation of the currency.
func (c Currency) String() string {
	return c.code
}

// Equals checks if two currencies are equal.
func (c Currency) Equals(other Currency) bool {
	return c.code == other.code
}

// IsZero checks if the currency is zero value.
func (c Currency) IsZero() bool {
	return c.code == ""
}

// Common currency constants
var (
	USD = MustNewCurrency("USD")
	EUR = MustNewCurrency("EUR")
	GBP = MustNewCurrency("GBP")
	JPY = MustNewCurrency("JPY")
	CAD = MustNewCurrency("CAD")
	AUD = MustNewCurrency("AUD")
	CHF = MustNewCurrency("CHF")
	CNY = MustNewCurrency("CNY")
	SEK = MustNewCurrency("SEK")
	NOK = MustNewCurrency("NOK")
	DKK = MustNewCurrency("DKK")
	PLN = MustNewCurrency("PLN")
	RUB = MustNewCurrency("RUB")
	INR = MustNewCurrency("INR")
	BRL = MustNewCurrency("BRL")
	KWD = MustNewCurrency("KWD")
	BHD = MustNewCurrency("BHD")
	JOD = MustNewCurrency("JOD")
)

// supportedCurrencies contains all supported currency codes.
var supportedCurrencies = map[string]bool{
	"USD": true, "EUR": true, "GBP": true, "JPY": true,
	"CAD": true, "AUD": true, "CHF": true, "CNY": true,
	"SEK": true, "NOK": true, "DKK": true, "PLN": true,
	"RUB": true, "INR": true, "BRL": true, "KWD": true,
	"BHD": true, "JOD": true,
}

// isValidCurrencyCode checks if the currency code is supported.
func isValidCurrencyCode(code string) bool {
	return supportedCurrencies[code]
}

// GetSupportedCurrencies returns a list of all supported currency codes.
func GetSupportedCurrencies() []string {
	currencies := make([]string, 0, len(supportedCurrencies))
	for code := range supportedCurrencies {
		currencies = append(currencies, code)
	}
	return currencies
}

// RegisterCurrency adds a new currency to the supported list.
// This is useful for adding custom or new currencies at runtime.
func RegisterCurrency(code string) error {
	code = strings.ToUpper(strings.TrimSpace(code))
	if len(code) != 3 {
		return ErrInvalidCurrencyCode
	}
	supportedCurrencies[code] = true
	return nil
}