package rules

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shortlink-org/go-sdk/specification"

	v1 "github.com/shortlink-org/billing/billing/internal/domain/account/v1"
)

// Domain-specific errors
var (
	ErrAccountIdRequired      = errors.New("account ID is required")
	ErrAccountInactive        = errors.New("account is not active")
	ErrAccountExpired         = errors.New("account has expired")
	ErrInvalidCreationDate    = errors.New("account creation date is invalid")
	ErrAccountStatusInvalid   = errors.New("account status is invalid")
)

// AccountIdRequired validates that the account has a valid ID
type AccountIdRequired struct{}

func NewAccountIdRequired() *AccountIdRequired {
	return &AccountIdRequired{}
}

func (s *AccountIdRequired) IsSatisfiedBy(account *v1.Account) error {
	if account.GetId() != uuid.Nil.String() && account.GetId() != "" {
		return nil
	}
	return ErrAccountIdRequired
}

// AccountActive validates that the account is active
type AccountActive struct{}

func NewAccountActive() *AccountActive {
	return &AccountActive{}
}

func (s *AccountActive) IsSatisfiedBy(account *v1.Account) error {
	if account.GetStatus() == v1.AccountStatus_ACCOUNT_STATUS_ACTIVE {
		return nil
	}
	return ErrAccountInactive
}

// AccountNotExpired validates that the account has not expired
type AccountNotExpired struct{}

func NewAccountNotExpired() *AccountNotExpired {
	return &AccountNotExpired{}
}

func (s *AccountNotExpired) IsSatisfiedBy(account *v1.Account) error {
	if account.GetExpiresAt() == nil {
		return nil // No expiration set
	}

	expiresAt := account.GetExpiresAt().AsTime()
	if time.Now().Before(expiresAt) {
		return nil
	}
	return ErrAccountExpired
}

// AccountCreatedAfter validates that the account was created after a specific date
type AccountCreatedAfter struct {
	MinDate time.Time
}

func NewAccountCreatedAfter(minDate time.Time) *AccountCreatedAfter {
	return &AccountCreatedAfter{MinDate: minDate}
}

func (s *AccountCreatedAfter) IsSatisfiedBy(account *v1.Account) error {
	if account.GetCreatedAt() == nil {
		return ErrInvalidCreationDate
	}

	createdAt := account.GetCreatedAt().AsTime()
	if createdAt.After(s.MinDate) {
		return nil
	}
	return fmt.Errorf("account created %v is before required date %v", createdAt, s.MinDate)
}

// AccountStatusIn validates that the account status is in the allowed list
type AccountStatusIn struct {
	AllowedStatuses []v1.AccountStatus
}

func NewAccountStatusIn(statuses ...v1.AccountStatus) *AccountStatusIn {
	return &AccountStatusIn{AllowedStatuses: statuses}
}

func (s *AccountStatusIn) IsSatisfiedBy(account *v1.Account) error {
	accountStatus := account.GetStatus()
	for _, status := range s.AllowedStatuses {
		if accountStatus == status {
			return nil
		}
	}
	return fmt.Errorf("account status %v is not in allowed statuses", accountStatus)
}

// Composite Specifications

// ValidAccountForBilling combines multiple account validations for billing operations
type ValidAccountForBilling struct {
	*specification.AndSpecification[v1.Account]
}

func NewValidAccountForBilling() *ValidAccountForBilling {
	andSpec := specification.NewAndSpecification[v1.Account](
		NewAccountIdRequired(),
		NewUserId(),
		NewTariffId(),
		NewAccountActive(),
		NewAccountNotExpired(),
	)
	
	return &ValidAccountForBilling{
		AndSpecification: andSpec,
	}
}

// ActiveAccountWithValidTariff ensures account is active and has a valid tariff
type ActiveAccountWithValidTariff struct {
	*specification.AndSpecification[v1.Account]
}

func NewActiveAccountWithValidTariff() *ActiveAccountWithValidTariff {
	andSpec := specification.NewAndSpecification[v1.Account](
		NewAccountActive(),
		NewTariffId(),
	)
	
	return &ActiveAccountWithValidTariff{
		AndSpecification: andSpec,
	}
}

// AccountSuspendedOrExpired checks if account is suspended OR expired
type AccountSuspendedOrExpired struct {
	*specification.OrSpecification[v1.Account]
}

func NewAccountSuspendedOrExpired() *AccountSuspendedOrExpired {
	suspendedSpec := NewAccountStatusIn(v1.AccountStatus_ACCOUNT_STATUS_SUSPENDED)
	expiredSpec := &specification.NotSpecification[v1.Account]{} // NOT(NotExpired) = Expired
	
	// We need to set up the NOT specification for expired accounts
	notExpiredSpec := NewAccountNotExpired()
	expiredSpecification := &struct {
		*specification.NotSpecification[v1.Account]
	}{}
	// Manually create the NOT specification since we need expired accounts
	expiredSpecification.NotSpecification = &specification.NotSpecification[v1.Account]{}
	
	orSpec := specification.NewOrSpecification[v1.Account](
		suspendedSpec,
		&AccountExpiredSpec{}, // Custom expired spec
	)
	
	return &AccountSuspendedOrExpired{
		OrSpecification: orSpec,
	}
}

// AccountExpiredSpec is a helper specification for expired accounts
type AccountExpiredSpec struct{}

func (s *AccountExpiredSpec) IsSatisfiedBy(account *v1.Account) error {
	notExpiredSpec := NewAccountNotExpired()
	err := notExpiredSpec.IsSatisfiedBy(account)
	if err == nil {
		return errors.New("account is not expired")
	}
	return nil // Account is expired, so this spec is satisfied
}

// Ensure all specifications implement the Specification interface
var (
	_ specification.Specification[v1.Account] = (*AccountIdRequired)(nil)
	_ specification.Specification[v1.Account] = (*AccountActive)(nil)
	_ specification.Specification[v1.Account] = (*AccountNotExpired)(nil)
	_ specification.Specification[v1.Account] = (*AccountCreatedAfter)(nil)
	_ specification.Specification[v1.Account] = (*AccountStatusIn)(nil)
	_ specification.Specification[v1.Account] = (*ValidAccountForBilling)(nil)
	_ specification.Specification[v1.Account] = (*ActiveAccountWithValidTariff)(nil)
	_ specification.Specification[v1.Account] = (*AccountSuspendedOrExpired)(nil)
	_ specification.Specification[v1.Account] = (*AccountExpiredSpec)(nil)
)