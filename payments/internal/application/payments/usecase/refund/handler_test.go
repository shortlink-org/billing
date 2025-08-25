package refund

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/type/money"

	"github.com/shortlink-org/billing/payments/internal/application/payments/ports"
	"github.com/shortlink-org/billing/payments/internal/application/payments/repository"
	eventv1 "github.com/shortlink-org/billing/payments/internal/domain/event/v1"
	flowv1 "github.com/shortlink-org/billing/payments/internal/domain/flow/v1"
	"github.com/shortlink-org/billing/payments/internal/domain/payment"
	"github.com/shortlink-org/billing/payments/internal/domain/payment/ledger"
)

// Mock implementations for testing
type MockPaymentRepository struct {
	mock.Mock
}

func (m *MockPaymentRepository) Save(ctx context.Context, p *payment.Payment, expectedVersion uint64) error {
	args := m.Called(ctx, p, expectedVersion)
	return args.Error(0)
}

func (m *MockPaymentRepository) Load(ctx context.Context, id uuid.UUID) (*payment.Payment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*payment.Payment), args.Error(1)
}

type MockPaymentProvider struct {
	mock.Mock
}

func (m *MockPaymentProvider) CreatePayment(ctx context.Context, in ports.CreatePaymentIn) (ports.CreatePaymentOut, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(ports.CreatePaymentOut), args.Error(1)
}

func (m *MockPaymentProvider) RefundPayment(ctx context.Context, in ports.RefundPaymentIn) (ports.RefundPaymentOut, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(ports.RefundPaymentOut), args.Error(1)
}

// Test helper to create a money amount
func money(currency string, units int64, nanos int32) *money.Money {
	return &money.Money{
		CurrencyCode: currency,
		Units:        units,
		Nanos:        nanos,
	}
}

// Test helper to create a paid payment aggregate
func createPaidPayment(t *testing.T, paymentID, invoiceID uuid.UUID, amount *money.Money) *payment.Payment {
	// Create a new payment
	agg, err := payment.New(paymentID, invoiceID, amount, eventv1.PaymentKind_PAYMENT_KIND_ONE_TIME, eventv1.CaptureMode_CAPTURE_MODE_AUTO)
	require.NoError(t, err)
	
	// Capture the full amount to make it paid
	err = agg.Capture(context.Background(), amount)
	require.NoError(t, err)
	
	return agg
}

func TestHandler_Handle_SuccessfulPartialRefund(t *testing.T) {
	// Arrange
	ctx := context.Background()
	paymentID := uuid.New()
	invoiceID := uuid.New()
	originalAmount := money("USD", 100, 0) // $100.00
	refundAmount := money("USD", 30, 0)    // $30.00
	
	agg := createPaidPayment(t, paymentID, invoiceID, originalAmount)
	
	mockRepo := &MockPaymentRepository{}
	mockProvider := &MockPaymentProvider{}
	
	mockRepo.On("Load", ctx, paymentID).Return(agg, nil)
	mockProvider.On("RefundPayment", ctx, mock.MatchedBy(func(in ports.RefundPaymentIn) bool {
		return in.PaymentID == paymentID && 
			   in.Amount.Units == 30 && 
			   in.Currency == "USD"
	})).Return(ports.RefundPaymentOut{
		Provider: ports.ProviderStripe,
		RefundID: "re_test123",
		Status:   ports.ProviderStatusSucceeded,
		Amount:   refundAmount,
	}, nil)
	mockRepo.On("Save", ctx, mock.AnythingOfType("*payment.Payment"), mock.AnythingOfType("uint64")).Return(nil)
	
	handler := &Handler{
		Repo:     mockRepo,
		Provider: mockProvider,
	}
	
	cmd := Command{
		PaymentID: paymentID,
		Amount:    refundAmount,
		Reason:    "customer request",
		Metadata:  map[string]string{"source": "test"},
	}
	
	// Act
	result, err := handler.Handle(ctx, cmd)
	
	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, paymentID, result.PaymentID)
	assert.Equal(t, "re_test123", result.RefundID)
	assert.Equal(t, refundAmount.Units, result.RefundAmount.Units)
	assert.Equal(t, refundAmount.Units, result.TotalRefunded.Units)
	assert.False(t, result.IsFullRefund)
	assert.Equal(t, flowv1.PaymentFlow_PAYMENT_FLOW_PAID, result.State)
	
	mockRepo.AssertExpectations(t)
	mockProvider.AssertExpectations(t)
}

func TestHandler_Handle_SuccessfulFullRefund(t *testing.T) {
	// Arrange
	ctx := context.Background()
	paymentID := uuid.New()
	invoiceID := uuid.New()
	originalAmount := money("USD", 100, 0) // $100.00
	
	agg := createPaidPayment(t, paymentID, invoiceID, originalAmount)
	
	mockRepo := &MockPaymentRepository{}
	mockProvider := &MockPaymentProvider{}
	
	mockRepo.On("Load", ctx, paymentID).Return(agg, nil)
	mockProvider.On("RefundPayment", ctx, mock.MatchedBy(func(in ports.RefundPaymentIn) bool {
		return in.PaymentID == paymentID && 
			   in.Amount.Units == 100 && 
			   in.Currency == "USD"
	})).Return(ports.RefundPaymentOut{
		Provider: ports.ProviderStripe,
		RefundID: "re_test123",
		Status:   ports.ProviderStatusSucceeded,
		Amount:   originalAmount,
	}, nil)
	mockRepo.On("Save", ctx, mock.AnythingOfType("*payment.Payment"), mock.AnythingOfType("uint64")).Return(nil)
	
	handler := &Handler{
		Repo:     mockRepo,
		Provider: mockProvider,
	}
	
	// Command with nil amount means full refund
	cmd := Command{
		PaymentID: paymentID,
		Amount:    nil, // Full refund
		Reason:    "order cancelled",
	}
	
	// Act
	result, err := handler.Handle(ctx, cmd)
	
	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, paymentID, result.PaymentID)
	assert.Equal(t, "re_test123", result.RefundID)
	assert.Equal(t, originalAmount.Units, result.RefundAmount.Units)
	assert.Equal(t, originalAmount.Units, result.TotalRefunded.Units)
	assert.True(t, result.IsFullRefund)
	assert.Equal(t, flowv1.PaymentFlow_PAYMENT_FLOW_REFUNDED, result.State)
	
	mockRepo.AssertExpectations(t)
	mockProvider.AssertExpectations(t)
}

func TestHandler_Handle_PaymentNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	paymentID := uuid.New()
	
	mockRepo := &MockPaymentRepository{}
	mockProvider := &MockPaymentProvider{}
	
	mockRepo.On("Load", ctx, paymentID).Return(nil, repository.ErrNotFound)
	
	handler := &Handler{
		Repo:     mockRepo,
		Provider: mockProvider,
	}
	
	cmd := Command{
		PaymentID: paymentID,
		Amount:    money("USD", 50, 0),
		Reason:    "test",
	}
	
	// Act
	result, err := handler.Handle(ctx, cmd)
	
	// Assert
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrPaymentNotFound)
	
	mockRepo.AssertExpectations(t)
}

func TestHandler_Handle_InvalidPaymentID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	
	mockRepo := &MockPaymentRepository{}
	mockProvider := &MockPaymentProvider{}
	
	handler := &Handler{
		Repo:     mockRepo,
		Provider: mockProvider,
	}
	
	cmd := Command{
		PaymentID: uuid.Nil, // Invalid payment ID
		Amount:    money("USD", 50, 0),
		Reason:    "test",
	}
	
	// Act
	result, err := handler.Handle(ctx, cmd)
	
	// Assert
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidRefundAmount)
}

func TestHandler_Handle_PaymentNotRefundable(t *testing.T) {
	// Arrange
	ctx := context.Background()
	paymentID := uuid.New()
	invoiceID := uuid.New()
	originalAmount := money("USD", 100, 0)
	
	// Create a payment in CREATED state (not refundable)
	agg, err := payment.New(paymentID, invoiceID, originalAmount, eventv1.PaymentKind_PAYMENT_KIND_ONE_TIME, eventv1.CaptureMode_CAPTURE_MODE_AUTO)
	require.NoError(t, err)
	
	mockRepo := &MockPaymentRepository{}
	mockProvider := &MockPaymentProvider{}
	
	mockRepo.On("Load", ctx, paymentID).Return(agg, nil)
	
	handler := &Handler{
		Repo:     mockRepo,
		Provider: mockProvider,
	}
	
	cmd := Command{
		PaymentID: paymentID,
		Amount:    money("USD", 50, 0),
		Reason:    "test",
	}
	
	// Act
	result, err := handler.Handle(ctx, cmd)
	
	// Assert
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrPaymentNotRefundable)
	
	mockRepo.AssertExpectations(t)
}

func TestHandler_Handle_InvalidRefundAmount(t *testing.T) {
	// Arrange
	ctx := context.Background()
	paymentID := uuid.New()
	invoiceID := uuid.New()
	originalAmount := money("USD", 100, 0)
	
	agg := createPaidPayment(t, paymentID, invoiceID, originalAmount)
	
	mockRepo := &MockPaymentRepository{}
	mockProvider := &MockPaymentProvider{}
	
	mockRepo.On("Load", ctx, paymentID).Return(agg, nil)
	
	handler := &Handler{
		Repo:     mockRepo,
		Provider: mockProvider,
	}
	
	testCases := []struct {
		name   string
		amount *money.Money
	}{
		{"zero amount", money("USD", 0, 0)},
		{"negative amount", money("USD", -10, 0)},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := Command{
				PaymentID: paymentID,
				Amount:    tc.amount,
				Reason:    "test",
			}
			
			// Act
			result, err := handler.Handle(ctx, cmd)
			
			// Assert
			assert.Nil(t, result)
			assert.Error(t, err)
			assert.ErrorIs(t, err, ErrInvalidRefundAmount)
		})
	}
	
	mockRepo.AssertExpectations(t)
}

func TestHandler_Handle_ProviderRefundFails(t *testing.T) {
	// Arrange
	ctx := context.Background()
	paymentID := uuid.New()
	invoiceID := uuid.New()
	originalAmount := money("USD", 100, 0)
	refundAmount := money("USD", 50, 0)
	
	agg := createPaidPayment(t, paymentID, invoiceID, originalAmount)
	
	mockRepo := &MockPaymentRepository{}
	mockProvider := &MockPaymentProvider{}
	
	mockRepo.On("Load", ctx, paymentID).Return(agg, nil)
	mockProvider.On("RefundPayment", ctx, mock.AnythingOfType("ports.RefundPaymentIn")).Return(
		ports.RefundPaymentOut{}, errors.New("provider error"))
	mockRepo.On("Save", ctx, mock.AnythingOfType("*payment.Payment"), mock.AnythingOfType("uint64")).Return(nil)
	
	handler := &Handler{
		Repo:     mockRepo,
		Provider: mockProvider,
	}
	
	cmd := Command{
		PaymentID: paymentID,
		Amount:    refundAmount,
		Reason:    "test",
	}
	
	// Act
	result, err := handler.Handle(ctx, cmd)
	
	// Assert
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider refund failed")
	
	mockRepo.AssertExpectations(t)
	mockProvider.AssertExpectations(t)
}

func TestHandler_Handle_AlreadyFullyRefunded(t *testing.T) {
	// Arrange
	ctx := context.Background()
	paymentID := uuid.New()
	invoiceID := uuid.New()
	originalAmount := money("USD", 100, 0)
	
	agg := createPaidPayment(t, paymentID, invoiceID, originalAmount)
	
	// Refund the full amount first
	_, err := agg.Refund(ctx, originalAmount)
	require.NoError(t, err)
	
	mockRepo := &MockPaymentRepository{}
	mockProvider := &MockPaymentProvider{}
	
	mockRepo.On("Load", ctx, paymentID).Return(agg, nil)
	
	handler := &Handler{
		Repo:     mockRepo,
		Provider: mockProvider,
	}
	
	// Try to refund again (full refund with nil amount)
	cmd := Command{
		PaymentID: paymentID,
		Amount:    nil, // Full refund
		Reason:    "test",
	}
	
	// Act
	result, err := handler.Handle(ctx, cmd)
	
	// Assert
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidRefundAmount)
	assert.Contains(t, err.Error(), "already fully refunded")
	
	mockRepo.AssertExpectations(t)
}

// Test helper functions
func TestIsZero(t *testing.T) {
	testCases := []struct {
		name     string
		money    *money.Money
		expected bool
	}{
		{"nil money", nil, true},
		{"zero money", money("USD", 0, 0), true},
		{"positive money", money("USD", 10, 0), false},
		{"negative money", money("USD", -10, 0), false},
		{"zero units with nanos", money("USD", 0, 100), false},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isZero(tc.money)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsNegative(t *testing.T) {
	testCases := []struct {
		name     string
		money    *money.Money
		expected bool
	}{
		{"nil money", nil, false},
		{"positive money", money("USD", 10, 0), false},
		{"zero money", money("USD", 0, 0), false},
		{"negative units", money("USD", -10, 0), true},
		{"zero units negative nanos", money("USD", 0, -100), true},
		{"negative units positive nanos", money("USD", -1, 100), true},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isNegative(tc.money)
			assert.Equal(t, tc.expected, result)
		})
	}
}