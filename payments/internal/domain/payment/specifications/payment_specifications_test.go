package specifications

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shortlink-org/go-sdk/specification"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/genproto/googleapis/type/money"

	eventv1 "github.com/shortlink-org/billing/payments/internal/domain/event/v1"
	flowv1 "github.com/shortlink-org/billing/payments/internal/domain/flow/v1"
	"github.com/shortlink-org/billing/payments/internal/domain/payment"
)

// PaymentSpecificationTestSuite groups all payment specification tests
type PaymentSpecificationTestSuite struct {
	suite.Suite
	validPayment   *payment.Payment
	invalidPayment *payment.Payment
	usdAmount      *money.Money
	eurAmount      *money.Money
}

func (suite *PaymentSpecificationTestSuite) SetupTest() {
	// Create valid payment
	validID := uuid.New()
	validInvoiceID := uuid.New()
	
	suite.usdAmount = &money.Money{
		CurrencyCode: "USD",
		Units:        100,
		Nanos:        0,
	}
	
	suite.eurAmount = &money.Money{
		CurrencyCode: "EUR",
		Units:        50,
		Nanos:        0,
	}
	
	var err error
	suite.validPayment, err = payment.New(
		validID,
		validInvoiceID,
		suite.usdAmount,
		eventv1.PaymentKind_PAYMENT_KIND_ONE_TIME,
		eventv1.CaptureMode_CAPTURE_MODE_MANUAL,
	)
	require.NoError(suite.T(), err)
	
	// Create invalid payment scenario (we'll use a different approach since we can't create truly invalid payments easily)
	// We'll test with nil values and edge cases in individual tests
}

func TestPaymentSpecificationSuite(t *testing.T) {
	suite.Run(t, new(PaymentSpecificationTestSuite))
}

// Test individual specifications

func (suite *PaymentSpecificationTestSuite) TestPaymentIdRequired() {
	spec := NewPaymentIdRequired()
	
	// Valid payment should pass
	err := spec.IsSatisfiedBy(suite.validPayment)
	assert.NoError(suite.T(), err)
	
	// Test would require creating payment with nil ID, which is not possible with current constructor
	// This demonstrates the value of the specification pattern - it can validate edge cases
}

func (suite *PaymentSpecificationTestSuite) TestInvoiceIdRequired() {
	spec := NewInvoiceIdRequired()
	
	// Valid payment should pass
	err := spec.IsSatisfiedBy(suite.validPayment)
	assert.NoError(suite.T(), err)
}

func (suite *PaymentSpecificationTestSuite) TestAmountRequired() {
	spec := NewAmountRequired()
	
	// Valid payment should pass
	err := spec.IsSatisfiedBy(suite.validPayment)
	assert.NoError(suite.T(), err)
}

func (suite *PaymentSpecificationTestSuite) TestAmountPositive() {
	spec := NewAmountPositive()
	
	// Valid payment should pass
	err := spec.IsSatisfiedBy(suite.validPayment)
	assert.NoError(suite.T(), err)
}

func (suite *PaymentSpecificationTestSuite) TestCurrencySupported() {
	// Test with supported currencies
	spec := NewCurrencySupported("USD", "EUR")
	
	err := spec.IsSatisfiedBy(suite.validPayment)
	assert.NoError(suite.T(), err)
	
	// Test with no supported currencies (all allowed)
	specAllAllowed := NewCurrencySupported()
	err = specAllAllowed.IsSatisfiedBy(suite.validPayment)
	assert.NoError(suite.T(), err)
	
	// Test with unsupported currency
	specUnsupported := NewCurrencySupported("GBP", "JPY")
	err = specUnsupported.IsSatisfiedBy(suite.validPayment)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "currency is not supported")
	assert.Contains(suite.T(), err.Error(), "USD")
}

func (suite *PaymentSpecificationTestSuite) TestPaymentInState() {
	// Test created state
	spec := NewPaymentInState(flowv1.PaymentFlow_PAYMENT_FLOW_CREATED)
	err := spec.IsSatisfiedBy(suite.validPayment)
	assert.NoError(suite.T(), err)
	
	// Test different state
	specAuth := NewPaymentInState(flowv1.PaymentFlow_PAYMENT_FLOW_AUTHORIZED)
	err = specAuth.IsSatisfiedBy(suite.validPayment)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "payment is in state")
	assert.Contains(suite.T(), err.Error(), "expected")
}

func (suite *PaymentSpecificationTestSuite) TestAmountWithinRange() {
	minAmount := &money.Money{CurrencyCode: "USD", Units: 10, Nanos: 0}
	maxAmount := &money.Money{CurrencyCode: "USD", Units: 1000, Nanos: 0}
	
	spec := NewAmountWithinRange(minAmount, maxAmount)
	
	// Valid payment (100 USD) should be within range
	err := spec.IsSatisfiedBy(suite.validPayment)
	assert.NoError(suite.T(), err)
	
	// Test with amount below minimum
	lowMinAmount := &money.Money{CurrencyCode: "USD", Units: 200, Nanos: 0}
	specLowMin := NewAmountWithinRange(lowMinAmount, maxAmount)
	err = specLowMin.IsSatisfiedBy(suite.validPayment)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "below minimum")
	
	// Test with amount above maximum  
	lowMaxAmount := &money.Money{CurrencyCode: "USD", Units: 50, Nanos: 0}
	specHighMax := NewAmountWithinRange(minAmount, lowMaxAmount)
	err = specHighMax.IsSatisfiedBy(suite.validPayment)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "exceeds maximum")
}

// Test composite specifications

func (suite *PaymentSpecificationTestSuite) TestValidPaymentForCreation() {
	spec := NewValidPaymentForCreation("USD", "EUR")
	
	// Valid payment should pass all checks
	err := spec.IsSatisfiedBy(suite.validPayment)
	assert.NoError(suite.T(), err)
	
	// Test with unsupported currency
	specUnsupported := NewValidPaymentForCreation("GBP")
	err = specUnsupported.IsSatisfiedBy(suite.validPayment)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "currency is not supported")
}

func (suite *PaymentSpecificationTestSuite) TestPaymentCanBeAuthorized() {
	spec := NewPaymentCanBeAuthorized()
	
	// Payment in created state should be authorizable
	err := spec.IsSatisfiedBy(suite.validPayment)
	assert.NoError(suite.T(), err)
}

func (suite *PaymentSpecificationTestSuite) TestPaymentCanBeCaptured() {
	spec := NewPaymentCanBeCaptured()
	
	// Payment in created state should be capturable (depending on policy)
	err := spec.IsSatisfiedBy(suite.validPayment)
	assert.NoError(suite.T(), err)
}

func (suite *PaymentSpecificationTestSuite) TestPaymentCanBeRefunded() {
	spec := NewPaymentCanBeRefunded()
	
	// Payment in created state cannot be refunded (needs to be paid first)
	err := spec.IsSatisfiedBy(suite.validPayment)
	assert.Error(suite.T(), err)
}

func (suite *PaymentSpecificationTestSuite) TestPaymentForFraudCheck() {
	highAmount := &money.Money{CurrencyCode: "USD", Units: 50, Nanos: 0}
	spec := NewPaymentForFraudCheck(highAmount, "USD")
	
	// Payment with amount 100 USD should trigger fraud check (above 50 USD threshold)
	err := spec.IsSatisfiedBy(suite.validPayment)
	assert.NoError(suite.T(), err) // No error means fraud check is required
	
	// Test with low amount and non-risky currency
	lowAmount := &money.Money{CurrencyCode: "USD", Units: 200, Nanos: 0}
	specLowRisk := NewPaymentForFraudCheck(lowAmount, "EUR")
	err = specLowRisk.IsSatisfiedBy(suite.validPayment)
	assert.Error(suite.T(), err) // Error means no fraud check needed
}

func (suite *PaymentSpecificationTestSuite) TestAmountAboveThreshold() {
	threshold := &money.Money{CurrencyCode: "USD", Units: 50, Nanos: 0}
	spec := &AmountAboveThreshold{Threshold: threshold}
	
	// 100 USD should be above 50 USD threshold
	err := spec.IsSatisfiedBy(suite.validPayment)
	assert.NoError(suite.T(), err)
	
	// Test with higher threshold
	highThreshold := &money.Money{CurrencyCode: "USD", Units: 200, Nanos: 0}
	specHigh := &AmountAboveThreshold{Threshold: highThreshold}
	err = specHigh.IsSatisfiedBy(suite.validPayment)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "below threshold")
}

func (suite *PaymentSpecificationTestSuite) TestCurrencyInList() {
	// USD should be in the list
	spec := &CurrencyInList{
		Currencies: map[string]struct{}{
			"USD": {},
			"EUR": {},
		},
	}
	
	err := spec.IsSatisfiedBy(suite.validPayment)
	assert.NoError(suite.T(), err)
	
	// USD should not be in this list
	specNoUSD := &CurrencyInList{
		Currencies: map[string]struct{}{
			"GBP": {},
			"JPY": {},
		},
	}
	
	err = specNoUSD.IsSatisfiedBy(suite.validPayment)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "not in the specified list")
}

// Test specification filtering

func (suite *PaymentSpecificationTestSuite) TestFilterPaymentsBySpecification() {
	// Create another payment for testing
	anotherPayment, err := payment.New(
		uuid.New(),
		uuid.New(),
		suite.eurAmount,
		eventv1.PaymentKind_PAYMENT_KIND_ONE_TIME,
		eventv1.CaptureMode_CAPTURE_MODE_IMMEDIATE,
	)
	require.NoError(suite.T(), err)
	
	payments := []*payment.Payment{suite.validPayment, anotherPayment}
	
	// Filter by currency
	spec := NewCurrencySupported("USD")
	filtered, err := FilterPaymentsBySpecification(payments, spec)
	
	// Should have error because EUR payment doesn't match
	assert.Error(suite.T(), err)
	// But should return the USD payment that passed
	assert.Len(suite.T(), filtered, 1)
	assert.Equal(suite.T(), suite.validPayment.ID(), filtered[0].ID())
}

// Test SpecificationBasedPolicy

func (suite *PaymentSpecificationTestSuite) TestSpecificationBasedPolicy() {
	policy := NewSpecificationBasedPolicy(
		WithSupportedCurrencies("USD", "EUR"),
		WithForceSCA(false),
	)
	
	// Test currency support
	assert.True(suite.T(), policy.IsCurrencySupported("USD"))
	assert.True(suite.T(), policy.IsCurrencySupported("EUR"))
	assert.False(suite.T(), policy.IsCurrencySupported("GBP"))
	
	// Test immediate capture
	assert.True(suite.T(), policy.AllowImmediateCapture(
		eventv1.PaymentKind_PAYMENT_KIND_ONE_TIME,
		eventv1.CaptureMode_CAPTURE_MODE_IMMEDIATE,
	))
	assert.False(suite.T(), policy.AllowImmediateCapture(
		eventv1.PaymentKind_PAYMENT_KIND_ONE_TIME,
		eventv1.CaptureMode_CAPTURE_MODE_MANUAL,
	))
	
	// Test validation methods
	err := policy.ValidatePaymentForCreation(suite.validPayment)
	assert.NoError(suite.T(), err)
	
	err = policy.ValidatePaymentForAuthorization(suite.validPayment)
	assert.NoError(suite.T(), err)
	
	err = policy.ValidatePaymentForCapture(suite.validPayment)
	assert.NoError(suite.T(), err)
	
	err = policy.ValidatePaymentForRefund(suite.validPayment)
	assert.Error(suite.T(), err) // Can't refund payment in created state
}

func (suite *PaymentSpecificationTestSuite) TestSpecificationBasedPolicy_WithOptions() {
	highAmount := &money.Money{CurrencyCode: "USD", Units: 1000, Nanos: 0}
	
	policy := NewSpecificationBasedPolicy(
		WithSupportedCurrencies("USD"),
		WithForceSCA(true),
		WithMinAmountForSCA(highAmount),
		WithHighRiskCurrencies("USD"),
	)
	
	// SCA should be forced
	assert.True(suite.T(), policy.ShouldRequireSCA(
		eventv1.PaymentKind_PAYMENT_KIND_ONE_TIME,
		eventv1.CaptureMode_CAPTURE_MODE_MANUAL,
	))
	
	// Should require fraud check for USD payments
	requiresFraudCheck := policy.RequiresFraudCheck(suite.validPayment)
	assert.True(suite.T(), requiresFraudCheck)
}

// Test helper functions

func (suite *PaymentSpecificationTestSuite) TestGetPaymentsReadyForCapture() {
	payments := []*payment.Payment{suite.validPayment}
	
	ready, err := GetPaymentsReadyForCapture(payments)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), ready, 1)
}

func (suite *PaymentSpecificationTestSuite) TestGetPaymentsRequiringFraudCheck() {
	policy := NewSpecificationBasedPolicy(
		WithMinAmountForSCA(&money.Money{CurrencyCode: "USD", Units: 50, Nanos: 0}),
		WithHighRiskCurrencies("USD"),
	)
	
	payments := []*payment.Payment{suite.validPayment}
	
	requiring, err := GetPaymentsRequiringFraudCheck(payments, policy)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), requiring, 1)
}

func (suite *PaymentSpecificationTestSuite) TestGetValidPaymentsForRefund() {
	payments := []*payment.Payment{suite.validPayment}
	
	valid, err := GetValidPaymentsForRefund(payments)
	// Should have error because created payments can't be refunded
	assert.Error(suite.T(), err)
	assert.Len(suite.T(), valid, 0)
}

// Benchmark tests

func BenchmarkPaymentSpecification_Simple(b *testing.B) {
	payment, _ := payment.New(
		uuid.New(),
		uuid.New(),
		&money.Money{CurrencyCode: "USD", Units: 100, Nanos: 0},
		eventv1.PaymentKind_PAYMENT_KIND_ONE_TIME,
		eventv1.CaptureMode_CAPTURE_MODE_MANUAL,
	)
	spec := NewAmountPositive()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = spec.IsSatisfiedBy(payment)
	}
}

func BenchmarkPaymentSpecification_Composite(b *testing.B) {
	payment, _ := payment.New(
		uuid.New(),
		uuid.New(),
		&money.Money{CurrencyCode: "USD", Units: 100, Nanos: 0},
		eventv1.PaymentKind_PAYMENT_KIND_ONE_TIME,
		eventv1.CaptureMode_CAPTURE_MODE_MANUAL,
	)
	spec := NewValidPaymentForCreation("USD", "EUR")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = spec.IsSatisfiedBy(payment)
	}
}

func BenchmarkPaymentSpecification_Filter(b *testing.B) {
	payments := make([]*payment.Payment, 1000)
	for i := range payments {
		payments[i], _ = payment.New(
			uuid.New(),
			uuid.New(),
			&money.Money{CurrencyCode: "USD", Units: 100, Nanos: 0},
			eventv1.PaymentKind_PAYMENT_KIND_ONE_TIME,
			eventv1.CaptureMode_CAPTURE_MODE_MANUAL,
		)
	}
	spec := NewAmountPositive()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = FilterPaymentsBySpecification(payments, spec)
	}
}

// Edge case tests

func TestPaymentSpecification_NilPayment(t *testing.T) {
	spec := NewAmountPositive()
	
	assert.Panics(t, func() {
		_ = spec.IsSatisfiedBy(nil)
	})
}

func TestPaymentSpecification_EmptySpecification(t *testing.T) {
	payment, err := payment.New(
		uuid.New(),
		uuid.New(),
		&money.Money{CurrencyCode: "USD", Units: 100, Nanos: 0},
		eventv1.PaymentKind_PAYMENT_KIND_ONE_TIME,
		eventv1.CaptureMode_CAPTURE_MODE_MANUAL,
	)
	require.NoError(t, err)
	
	// Test empty AND specification
	andSpec := specification.NewAndSpecification[payment.Payment]()
	err = andSpec.IsSatisfiedBy(payment)
	assert.NoError(t, err)
	
	// Test empty OR specification
	orSpec := specification.NewOrSpecification[payment.Payment]()
	err = orSpec.IsSatisfiedBy(payment)
	assert.Error(t, err) // OR with no specs should fail
}

// Integration tests

func TestPaymentSpecification_ComplexBusinessRule(t *testing.T) {
	// Business rule: High-value card payments in USD require fraud checking
	highValueUSDCardSpec := specification.NewAndSpecification[payment.Payment](
		NewPaymentInState(flowv1.PaymentFlow_PAYMENT_FLOW_CREATED),
		&AmountAboveThreshold{
			Threshold: &money.Money{CurrencyCode: "USD", Units: 500, Nanos: 0},
		},
		&CurrencyInList{
			Currencies: map[string]struct{}{"USD": {}},
		},
	)
	
	// Create high-value USD payment
	highValuePayment, err := payment.New(
		uuid.New(),
		uuid.New(),
		&money.Money{CurrencyCode: "USD", Units: 1000, Nanos: 0},
		eventv1.PaymentKind_PAYMENT_KIND_ONE_TIME,
		eventv1.CaptureMode_CAPTURE_MODE_MANUAL,
	)
	require.NoError(t, err)
	
	// Should satisfy the complex rule
	err = highValueUSDCardSpec.IsSatisfiedBy(highValuePayment)
	assert.NoError(t, err)
	
	// Create low-value payment
	lowValuePayment, err := payment.New(
		uuid.New(),
		uuid.New(),
		&money.Money{CurrencyCode: "USD", Units: 100, Nanos: 0},
		eventv1.PaymentKind_PAYMENT_KIND_ONE_TIME,
		eventv1.CaptureMode_CAPTURE_MODE_MANUAL,
	)
	require.NoError(t, err)
	
	// Should not satisfy the complex rule
	err = highValueUSDCardSpec.IsSatisfiedBy(lowValuePayment)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "below threshold")
}