package rules

import (
	"fmt"

	"github.com/shortlink-org/go-sdk/specification"

	v1 "github.com/shortlink-org/billing/billing/internal/domain/account/v1"
)

// AccountService demonstrates using specifications in domain services
type AccountService struct {
	// Dependencies would be injected here
}

// NewAccountService creates a new account service
func NewAccountService() *AccountService {
	return &AccountService{}
}

// ValidateAccountForBilling validates an account using composite specifications
func (s *AccountService) ValidateAccountForBilling(account *v1.Account) error {
	spec := NewValidAccountForBilling()
	return spec.IsSatisfiedBy(account)
}

// ValidateAccountForActivation validates an account can be activated
func (s *AccountService) ValidateAccountForActivation(account *v1.Account) error {
	// Account must have user ID and tariff ID to be activated
	spec := specification.NewAndSpecification[v1.Account](
		NewUserId(),
		NewTariffId(),
		NewAccountIdRequired(),
	)
	return spec.IsSatisfiedBy(account)
}

// CanAccountBeDeactivated checks if an account can be deactivated
func (s *AccountService) CanAccountBeDeactivated(account *v1.Account) bool {
	// Account can be deactivated if it's currently active
	spec := NewAccountActive()
	return spec.IsSatisfiedBy(account) == nil
}

// GetAccountsEligibleForUpgrade returns accounts that are eligible for upgrade
func (s *AccountService) GetAccountsEligibleForUpgrade(accounts []*v1.Account) ([]*v1.Account, error) {
	// Accounts eligible for upgrade: active accounts with valid user and tariff
	spec := specification.NewAndSpecification[v1.Account](
		NewAccountActive(),
		NewUserId(),
		NewTariffId(),
		NewAccountNotExpired(),
	)
	
	return specification.Filter(accounts, spec)
}

// GetSuspendedOrExpiredAccounts returns accounts that are suspended or expired
func (s *AccountService) GetSuspendedOrExpiredAccounts(accounts []*v1.Account) ([]*v1.Account, error) {
	spec := NewAccountSuspendedOrExpired()
	return specification.Filter(accounts, spec)
}

// GetActiveAccountsWithValidTariff returns active accounts with valid tariffs
func (s *AccountService) GetActiveAccountsWithValidTariff(accounts []*v1.Account) ([]*v1.Account, error) {
	spec := NewActiveAccountWithValidTariff()
	return specification.Filter(accounts, spec)
}

// ValidateMultipleAccounts validates multiple accounts and returns validation results
func (s *AccountService) ValidateMultipleAccounts(accounts []*v1.Account, spec specification.Specification[v1.Account]) map[string]error {
	results := make(map[string]error)
	
	for _, account := range accounts {
		accountID := account.GetId()
		if accountID == "" {
			results["unknown"] = NewAccountIdRequired().IsSatisfiedBy(account)
			continue
		}
		
		err := spec.IsSatisfiedBy(account)
		if err != nil {
			results[accountID] = err
		}
	}
	
	return results
}

// Example of complex business logic using specifications
func (s *AccountService) ProcessAccountForRenewal(account *v1.Account) error {
	// Step 1: Validate account is eligible for renewal
	eligibilitySpec := specification.NewAndSpecification[v1.Account](
		NewAccountIdRequired(),
		NewUserId(),
		NewTariffId(),
		// Account must be active OR expired (but not suspended)
		specification.NewOrSpecification[v1.Account](
			NewAccountActive(),
			&AccountExpiredSpec{},
		),
	)
	
	if err := eligibilitySpec.IsSatisfiedBy(account); err != nil {
		return fmt.Errorf("account not eligible for renewal: %w", err)
	}
	
	// Step 2: Additional business logic would go here
	// For example, checking payment history, calculating renewal cost, etc.
	
	return nil
}

// AccountSpecificationBuilder provides a fluent interface for building complex specifications
type AccountSpecificationBuilder struct {
	specs []specification.Specification[v1.Account]
}

// NewAccountSpecificationBuilder creates a new specification builder
func NewAccountSpecificationBuilder() *AccountSpecificationBuilder {
	return &AccountSpecificationBuilder{}
}

// WithUserId adds user ID requirement
func (b *AccountSpecificationBuilder) WithUserId() *AccountSpecificationBuilder {
	b.specs = append(b.specs, NewUserId())
	return b
}

// WithTariffId adds tariff ID requirement
func (b *AccountSpecificationBuilder) WithTariffId() *AccountSpecificationBuilder {
	b.specs = append(b.specs, NewTariffId())
	return b
}

// WithActiveStatus adds active status requirement
func (b *AccountSpecificationBuilder) WithActiveStatus() *AccountSpecificationBuilder {
	b.specs = append(b.specs, NewAccountActive())
	return b
}

// WithNotExpired adds not expired requirement
func (b *AccountSpecificationBuilder) WithNotExpired() *AccountSpecificationBuilder {
	b.specs = append(b.specs, NewAccountNotExpired())
	return b
}

// WithAccountId adds account ID requirement
func (b *AccountSpecificationBuilder) WithAccountId() *AccountSpecificationBuilder {
	b.specs = append(b.specs, NewAccountIdRequired())
	return b
}

// WithStatusIn adds status requirement
func (b *AccountSpecificationBuilder) WithStatusIn(statuses ...v1.AccountStatus) *AccountSpecificationBuilder {
	b.specs = append(b.specs, NewAccountStatusIn(statuses...))
	return b
}

// BuildAnd creates an AND specification from all added specs
func (b *AccountSpecificationBuilder) BuildAnd() specification.Specification[v1.Account] {
	if len(b.specs) == 0 {
		return &AlwaysPassSpecification{}
	}
	return specification.NewAndSpecification[v1.Account](b.specs...)
}

// BuildOr creates an OR specification from all added specs
func (b *AccountSpecificationBuilder) BuildOr() specification.Specification[v1.Account] {
	if len(b.specs) == 0 {
		return &AlwaysPassSpecification{}
	}
	return specification.NewOrSpecification[v1.Account](b.specs...)
}

// AlwaysPassSpecification is a helper specification that always passes
type AlwaysPassSpecification struct{}

func (s *AlwaysPassSpecification) IsSatisfiedBy(account *v1.Account) error {
	return nil
}

// Example usage:
// spec := NewAccountSpecificationBuilder().
//     WithUserId().
//     WithTariffId().
//     WithActiveStatus().
//     WithNotExpired().
//     BuildAnd()
//
// err := spec.IsSatisfiedBy(account)

// Domain specification combinations for common use cases

// GetBasicValidationSpec returns a specification for basic account validation
func GetBasicValidationSpec() specification.Specification[v1.Account] {
	return NewAccountSpecificationBuilder().
		WithAccountId().
		WithUserId().
		WithTariffId().
		BuildAnd()
}

// GetFullValidationSpec returns a specification for complete account validation
func GetFullValidationSpec() specification.Specification[v1.Account] {
	return NewAccountSpecificationBuilder().
		WithAccountId().
		WithUserId().
		WithTariffId().
		WithActiveStatus().
		WithNotExpired().
		BuildAnd()
}

// GetRenewalEligibilitySpec returns a specification for renewal eligibility
func GetRenewalEligibilitySpec() specification.Specification[v1.Account] {
	basicSpec := GetBasicValidationSpec()
	statusSpec := specification.NewOrSpecification[v1.Account](
		NewAccountActive(),
		&AccountExpiredSpec{},
	)
	
	return specification.NewAndSpecification[v1.Account](
		basicSpec,
		statusSpec,
	)
}