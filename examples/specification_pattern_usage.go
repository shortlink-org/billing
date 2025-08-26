// Package examples demonstrates the usage of the specification pattern in domain-level validation
package examples

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/shortlink-org/go-sdk/specification"
	"google.golang.org/genproto/googleapis/type/money"
	"google.golang.org/protobuf/types/known/timestamppb"

	// Billing domain specifications
	accountRules "github.com/shortlink-org/shortlink/boundaries/billing/billing/internal/domain/account/v1/rules"
	v1 "github.com/shortlink-org/shortlink/boundaries/billing/billing/internal/domain/account/v1"

	// Payment domain specifications
	eventv1 "github.com/shortlink-org/billing/payments/internal/domain/event/v1"
	"github.com/shortlink-org/billing/payments/internal/domain/payment"
	paymentSpecs "github.com/shortlink-org/billing/payments/internal/domain/payment/specifications"
)

// ExampleBasicSpecificationUsage demonstrates basic specification pattern usage
func ExampleBasicSpecificationUsage() {
	fmt.Println("=== Basic Specification Usage ===")
	
	// Create a sample account
	account := &v1.Account{
		Id:       uuid.New().String(),
		UserId:   uuid.New().String(),
		TariffId: uuid.New().String(),
		Status:   v1.AccountStatus_ACCOUNT_STATUS_ACTIVE,
		CreatedAt: timestamppb.New(time.Now()),
		ExpiresAt: timestamppb.New(time.Now().Add(30 * 24 * time.Hour)),
	}

	// 1. Simple specification validation
	userIdSpec := accountRules.NewUserId()
	if err := userIdSpec.IsSatisfiedBy(account); err != nil {
		log.Printf("User ID validation failed: %v", err)
	} else {
		fmt.Println("✓ User ID validation passed")
	}

	// 2. Composite specification validation
	billingSpec := accountRules.NewValidAccountForBilling()
	if err := billingSpec.IsSatisfiedBy(account); err != nil {
		log.Printf("Billing validation failed: %v", err)
	} else {
		fmt.Println("✓ Billing validation passed")
	}

	// 3. Builder pattern for specifications
	customSpec := accountRules.NewAccountSpecificationBuilder().
		WithUserId().
		WithTariffId().
		WithActiveStatus().
		WithNotExpired().
		BuildAnd()

	if err := customSpec.IsSatisfiedBy(account); err != nil {
		log.Printf("Custom validation failed: %v", err)
	} else {
		fmt.Println("✓ Custom validation passed")
	}
}

// ExampleCompositeSpecifications demonstrates AND, OR, NOT operations
func ExampleCompositeSpecifications() {
	fmt.Println("\n=== Composite Specifications ===")
	
	// Create test accounts
	activeAccount := &v1.Account{
		Id:       uuid.New().String(),
		UserId:   uuid.New().String(),
		TariffId: uuid.New().String(),
		Status:   v1.AccountStatus_ACCOUNT_STATUS_ACTIVE,
	}
	
	suspendedAccount := &v1.Account{
		Id:       uuid.New().String(),
		UserId:   uuid.New().String(),
		TariffId: uuid.New().String(),
		Status:   v1.AccountStatus_ACCOUNT_STATUS_SUSPENDED,
	}

	// 1. AND specification - all conditions must be met
	andSpec := specification.NewAndSpecification[v1.Account](
		accountRules.NewUserId(),
		accountRules.NewTariffId(),
		accountRules.NewAccountActive(),
	)

	fmt.Println("AND Specification Results:")
	fmt.Printf("  Active account: %v\n", andSpec.IsSatisfiedBy(activeAccount) == nil)
	fmt.Printf("  Suspended account: %v\n", andSpec.IsSatisfiedBy(suspendedAccount) == nil)

	// 2. OR specification - any condition can be met
	orSpec := specification.NewOrSpecification[v1.Account](
		accountRules.NewAccountActive(),
		accountRules.NewAccountStatusIn(v1.AccountStatus_ACCOUNT_STATUS_PENDING),
	)

	fmt.Println("OR Specification Results:")
	fmt.Printf("  Active account: %v\n", orSpec.IsSatisfiedBy(activeAccount) == nil)
	fmt.Printf("  Suspended account: %v\n", orSpec.IsSatisfiedBy(suspendedAccount) == nil)

	// 3. NOT specification - condition must NOT be met
	notActiveSpec := specification.NewNotSpecification[v1.Account](
		accountRules.NewAccountActive(),
	)

	fmt.Println("NOT Specification Results:")
	fmt.Printf("  Active account (NOT active): %v\n", notActiveSpec.IsSatisfiedBy(activeAccount) == nil)
	fmt.Printf("  Suspended account (NOT active): %v\n", notActiveSpec.IsSatisfiedBy(suspendedAccount) == nil)
}

// ExamplePaymentSpecifications demonstrates payment domain specifications
func ExamplePaymentSpecifications() {
	fmt.Println("\n=== Payment Specifications ===")
	
	// Create a sample payment
	paymentID := uuid.New()
	invoiceID := uuid.New()
	amount := &money.Money{
		CurrencyCode: "USD",
		Units:        100,
		Nanos:        50000000, // $100.50
	}

	payment, err := payment.New(
		paymentID,
		invoiceID,
		amount,
		eventv1.PaymentKind_PAYMENT_KIND_CARD,
		eventv1.CaptureMode_CAPTURE_MODE_MANUAL,
	)
	if err != nil {
		log.Fatalf("Failed to create payment: %v", err)
	}

	// 1. Basic payment validations
	specs := map[string]specification.Specification[payment.Payment]{
		"Payment ID Required": paymentSpecs.NewPaymentIdRequired(),
		"Invoice ID Required": paymentSpecs.NewInvoiceIdRequired(),
		"Amount Required":     paymentSpecs.NewAmountRequired(),
		"Amount Positive":     paymentSpecs.NewAmountPositive(),
		"Currency Supported":  paymentSpecs.NewCurrencySupported("USD", "EUR"),
	}

	fmt.Println("Basic Payment Validations:")
	for name, spec := range specs {
		result := spec.IsSatisfiedBy(payment)
		status := "✓ PASS"
		if result != nil {
			status = "✗ FAIL: " + result.Error()
		}
		fmt.Printf("  %s: %s\n", name, status)
	}

	// 2. Composite payment specifications
	fmt.Println("\nComposite Payment Validations:")
	
	creationSpec := paymentSpecs.NewValidPaymentForCreation("USD", "EUR")
	if err := creationSpec.IsSatisfiedBy(payment); err != nil {
		fmt.Printf("  Creation validation: ✗ FAIL: %v\n", err)
	} else {
		fmt.Println("  Creation validation: ✓ PASS")
	}

	authSpec := paymentSpecs.NewPaymentCanBeAuthorized()
	if err := authSpec.IsSatisfiedBy(payment); err != nil {
		fmt.Printf("  Authorization validation: ✗ FAIL: %v\n", err)
	} else {
		fmt.Println("  Authorization validation: ✓ PASS")
	}

	captureSpec := paymentSpecs.NewPaymentCanBeCaptured()
	if err := captureSpec.IsSatisfiedBy(payment); err != nil {
		fmt.Printf("  Capture validation: ✗ FAIL: %v\n", err)
	} else {
		fmt.Println("  Capture validation: ✓ PASS")
	}

	refundSpec := paymentSpecs.NewPaymentCanBeRefunded()
	if err := refundSpec.IsSatisfiedBy(payment); err != nil {
		fmt.Printf("  Refund validation: ✗ FAIL: %v\n", err)
	} else {
		fmt.Println("  Refund validation: ✓ PASS")
	}
}

// ExampleFilteringWithSpecifications demonstrates filtering collections
func ExampleFilteringWithSpecifications() {
	fmt.Println("\n=== Filtering with Specifications ===")
	
	// Create multiple accounts with different statuses
	accounts := []*v1.Account{
		{
			Id:       uuid.New().String(),
			UserId:   uuid.New().String(),
			TariffId: uuid.New().String(),
			Status:   v1.AccountStatus_ACCOUNT_STATUS_ACTIVE,
		},
		{
			Id:       uuid.New().String(),
			UserId:   uuid.New().String(),
			TariffId: "", // Missing tariff
			Status:   v1.AccountStatus_ACCOUNT_STATUS_ACTIVE,
		},
		{
			Id:       uuid.New().String(),
			UserId:   uuid.New().String(),
			TariffId: uuid.New().String(),
			Status:   v1.AccountStatus_ACCOUNT_STATUS_SUSPENDED,
		},
		{
			Id:       uuid.New().String(),
			UserId:   "", // Missing user ID
			TariffId: uuid.New().String(),
			Status:   v1.AccountStatus_ACCOUNT_STATUS_ACTIVE,
		},
	}

	fmt.Printf("Total accounts: %d\n", len(accounts))

	// Filter active accounts
	activeSpec := accountRules.NewAccountActive()
	activeAccounts, err := specification.Filter(accounts, activeSpec)
	if err != nil {
		fmt.Printf("Filtering active accounts had errors: %v\n", err)
	}
	fmt.Printf("Active accounts: %d\n", len(activeAccounts))

	// Filter accounts ready for billing
	service := accountRules.NewAccountService()
	billingReady, err := service.GetAccountsEligibleForUpgrade(accounts)
	if err != nil {
		fmt.Printf("Filtering billing-ready accounts had errors: %v\n", err)
	}
	fmt.Printf("Billing-ready accounts: %d\n", len(billingReady))

	// Filter with custom specification
	customSpec := specification.NewAndSpecification[v1.Account](
		accountRules.NewUserId(),
		accountRules.NewTariffId(),
	)
	validAccounts, err := specification.Filter(accounts, customSpec)
	if err != nil {
		fmt.Printf("Custom filtering had errors: %v\n", err)
	}
	fmt.Printf("Accounts with user and tariff IDs: %d\n", len(validAccounts))
}

// ExamplePolicyBasedValidation demonstrates policy-based validation using specifications
func ExamplePolicyBasedValidation() {
	fmt.Println("\n=== Policy-Based Validation ===")
	
	// Create a payment policy using specifications
	policy := paymentSpecs.NewSpecificationBasedPolicy(
		paymentSpecs.WithSupportedCurrencies("USD", "EUR", "GBP"),
		paymentSpecs.WithForceSCA(false),
		paymentSpecs.WithMinAmountForSCA(&money.Money{CurrencyCode: "USD", Units: 500}),
		paymentSpecs.WithHighRiskCurrencies("USD"),
	)

	// Create test payments
	lowValuePayment, _ := payment.New(
		uuid.New(), uuid.New(),
		&money.Money{CurrencyCode: "USD", Units: 100},
		eventv1.PaymentKind_PAYMENT_KIND_CARD,
		eventv1.CaptureMode_CAPTURE_MODE_MANUAL,
	)

	highValuePayment, _ := payment.New(
		uuid.New(), uuid.New(),
		&money.Money{CurrencyCode: "USD", Units: 1000},
		eventv1.PaymentKind_PAYMENT_KIND_CARD,
		eventv1.CaptureMode_CAPTURE_MODE_MANUAL,
	)

	unsupportedCurrencyPayment, _ := payment.New(
		uuid.New(), uuid.New(),
		&money.Money{CurrencyCode: "JPY", Units: 10000},
		eventv1.PaymentKind_PAYMENT_KIND_CARD,
		eventv1.CaptureMode_CAPTURE_MODE_MANUAL,
	)

	payments := []*payment.Payment{lowValuePayment, highValuePayment, unsupportedCurrencyPayment}

	fmt.Println("Policy Validation Results:")

	for i, p := range payments {
		fmt.Printf("\nPayment %d (%s %d):\n", i+1, p.Ledger.Amount.CurrencyCode, p.Ledger.Amount.Units)
		
		// Currency support
		supported := policy.IsCurrencySupported(p.Ledger.Amount.CurrencyCode)
		fmt.Printf("  Currency supported: %v\n", supported)

		// Creation validation
		if err := policy.ValidatePaymentForCreation(p); err != nil {
			fmt.Printf("  Creation validation: ✗ FAIL: %v\n", err)
		} else {
			fmt.Printf("  Creation validation: ✓ PASS\n")
		}

		// Fraud check requirement
		requiresFraudCheck := policy.RequiresFraudCheck(p)
		fmt.Printf("  Requires fraud check: %v\n", requiresFraudCheck)

		// Authorization validation
		if err := policy.ValidatePaymentForAuthorization(p); err != nil {
			fmt.Printf("  Authorization validation: ✗ FAIL: %v\n", err)
		} else {
			fmt.Printf("  Authorization validation: ✓ PASS\n")
		}
	}
}

// ExampleBusinessRuleComposition demonstrates complex business rule composition
func ExampleBusinessRuleComposition() {
	fmt.Println("\n=== Complex Business Rule Composition ===")
	
	// Business Rule: "Premium account renewal eligibility"
	// An account is eligible for premium renewal if:
	// 1. It has a valid user ID and tariff ID
	// 2. It is either active OR expired (but not suspended)
	// 3. It was created more than 30 days ago
	// 4. It is NOT in a trial status

	thirtyDaysAgo := time.Now().Add(-30 * 24 * time.Hour)
	
	premiumRenewalSpec := specification.NewAndSpecification[v1.Account](
		// Basic requirements
		specification.NewAndSpecification[v1.Account](
			accountRules.NewUserId(),
			accountRules.NewTariffId(),
		),
		// Status requirements (active OR expired, but not suspended)
		specification.NewAndSpecification[v1.Account](
			specification.NewOrSpecification[v1.Account](
				accountRules.NewAccountActive(),
				&accountRules.AccountExpiredSpec{},
			),
			specification.NewNotSpecification[v1.Account](
				accountRules.NewAccountStatusIn(v1.AccountStatus_ACCOUNT_STATUS_SUSPENDED),
			),
		),
		// Age requirement
		accountRules.NewAccountCreatedAfter(thirtyDaysAgo),
		// NOT trial status
		specification.NewNotSpecification[v1.Account](
			accountRules.NewAccountStatusIn(v1.AccountStatus_ACCOUNT_STATUS_PENDING),
		),
	)

	// Test accounts
	testAccounts := []*v1.Account{
		// Eligible account
		{
			Id:        uuid.New().String(),
			UserId:    uuid.New().String(),
			TariffId:  uuid.New().String(),
			Status:    v1.AccountStatus_ACCOUNT_STATUS_ACTIVE,
			CreatedAt: timestamppb.New(time.Now().Add(-60 * 24 * time.Hour)),
		},
		// Too new
		{
			Id:        uuid.New().String(),
			UserId:    uuid.New().String(),
			TariffId:  uuid.New().String(),
			Status:    v1.AccountStatus_ACCOUNT_STATUS_ACTIVE,
			CreatedAt: timestamppb.New(time.Now().Add(-10 * 24 * time.Hour)),
		},
		// Suspended
		{
			Id:        uuid.New().String(),
			UserId:    uuid.New().String(),
			TariffId:  uuid.New().String(),
			Status:    v1.AccountStatus_ACCOUNT_STATUS_SUSPENDED,
			CreatedAt: timestamppb.New(time.Now().Add(-60 * 24 * time.Hour)),
		},
	}

	fmt.Println("Premium Renewal Eligibility:")
	for i, account := range testAccounts {
		result := premiumRenewalSpec.IsSatisfiedBy(account)
		status := "✓ ELIGIBLE"
		if result != nil {
			status = "✗ NOT ELIGIBLE: " + result.Error()
		}
		fmt.Printf("  Account %d: %s\n", i+1, status)
	}
}

// ExampleSpecificationPerformance demonstrates performance considerations
func ExampleSpecificationPerformance() {
	fmt.Println("\n=== Performance Considerations ===")
	
	// Create a large number of accounts for performance testing
	const numAccounts = 10000
	accounts := make([]*v1.Account, numAccounts)
	
	for i := 0; i < numAccounts; i++ {
		accounts[i] = &v1.Account{
			Id:       uuid.New().String(),
			UserId:   uuid.New().String(),
			TariffId: uuid.New().String(),
			Status:   v1.AccountStatus_ACCOUNT_STATUS_ACTIVE,
		}
		
		// Make some accounts invalid for testing
		if i%10 == 0 {
			accounts[i].UserId = ""
		}
		if i%15 == 0 {
			accounts[i].Status = v1.AccountStatus_ACCOUNT_STATUS_SUSPENDED
		}
	}

	// Performance test: simple specification
	start := time.Now()
	simpleSpec := accountRules.NewUserId()
	validUsers, _ := specification.Filter(accounts, simpleSpec)
	simpleTime := time.Since(start)
	
	// Performance test: composite specification
	start = time.Now()
	compositeSpec := accountRules.NewValidAccountForBilling()
	validBilling, _ := specification.Filter(accounts, compositeSpec)
	compositeTime := time.Since(start)

	fmt.Printf("Performance results for %d accounts:\n", numAccounts)
	fmt.Printf("  Simple specification: %v (%d valid)\n", simpleTime, len(validUsers))
	fmt.Printf("  Composite specification: %v (%d valid)\n", compositeTime, len(validBilling))
	fmt.Printf("  Composite overhead: %.2fx\n", float64(compositeTime)/float64(simpleTime))
}

// RunAllExamples runs all specification pattern examples
func RunAllExamples() {
	fmt.Println("Specification Pattern Examples")
	fmt.Println("============================")
	
	ExampleBasicSpecificationUsage()
	ExampleCompositeSpecifications()
	ExamplePaymentSpecifications()
	ExampleFilteringWithSpecifications()
	ExamplePolicyBasedValidation()
	ExampleBusinessRuleComposition()
	ExampleSpecificationPerformance()
	
	fmt.Println("\n=== Examples Complete ===")
}