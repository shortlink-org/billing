package rules

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shortlink-org/go-sdk/specification"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "github.com/shortlink-org/billing/billing/internal/domain/account/v1"
)

// AccountSpecificationTestSuite groups all account specification tests
type AccountSpecificationTestSuite struct {
	suite.Suite
	validAccount   *v1.Account
	invalidAccount *v1.Account
}

func (suite *AccountSpecificationTestSuite) SetupTest() {
	now := time.Now()
	future := now.Add(time.Hour * 24 * 30) // 30 days from now
	
	suite.validAccount = &v1.Account{
		Id:        uuid.New().String(),
		UserId:    uuid.New().String(),
		TariffId:  uuid.New().String(),
		Status:    v1.AccountStatus_ACCOUNT_STATUS_ACTIVE,
		CreatedAt: timestamppb.New(now),
		ExpiresAt: timestamppb.New(future),
	}
	
	suite.invalidAccount = &v1.Account{
		Id:        "",
		UserId:    uuid.Nil.String(),
		TariffId:  uuid.Nil.String(),
		Status:    v1.AccountStatus_ACCOUNT_STATUS_SUSPENDED,
		CreatedAt: nil,
		ExpiresAt: timestamppb.New(now.Add(-time.Hour)), // Expired
	}
}

func TestAccountSpecificationSuite(t *testing.T) {
	suite.Run(t, new(AccountSpecificationTestSuite))
}

// Test individual specifications

func (suite *AccountSpecificationTestSuite) TestAccountIdRequired() {
	spec := NewAccountIdRequired()
	
	// Valid account should pass
	err := spec.IsSatisfiedBy(suite.validAccount)
	assert.NoError(suite.T(), err)
	
	// Invalid account should fail
	err = spec.IsSatisfiedBy(suite.invalidAccount)
	assert.Error(suite.T(), err)
	assert.ErrorIs(suite.T(), err, ErrAccountIdRequired)
}

func (suite *AccountSpecificationTestSuite) TestAccountActive() {
	spec := NewAccountActive()
	
	// Valid (active) account should pass
	err := spec.IsSatisfiedBy(suite.validAccount)
	assert.NoError(suite.T(), err)
	
	// Invalid (suspended) account should fail
	err = spec.IsSatisfiedBy(suite.invalidAccount)
	assert.Error(suite.T(), err)
	assert.ErrorIs(suite.T(), err, ErrAccountInactive)
}

func (suite *AccountSpecificationTestSuite) TestAccountNotExpired() {
	spec := NewAccountNotExpired()
	
	// Valid (not expired) account should pass
	err := spec.IsSatisfiedBy(suite.validAccount)
	assert.NoError(suite.T(), err)
	
	// Invalid (expired) account should fail
	err = spec.IsSatisfiedBy(suite.invalidAccount)
	assert.Error(suite.T(), err)
	assert.ErrorIs(suite.T(), err, ErrAccountExpired)
	
	// Account with no expiration should pass
	accountNoExpiry := &v1.Account{
		Id:        uuid.New().String(),
		Status:    v1.AccountStatus_ACCOUNT_STATUS_ACTIVE,
		ExpiresAt: nil,
	}
	err = spec.IsSatisfiedBy(accountNoExpiry)
	assert.NoError(suite.T(), err)
}

func (suite *AccountSpecificationTestSuite) TestAccountCreatedAfter() {
	yesterday := time.Now().Add(-24 * time.Hour)
	tomorrow := time.Now().Add(24 * time.Hour)
	
	spec := NewAccountCreatedAfter(yesterday)
	
	// Account created after the date should pass
	err := spec.IsSatisfiedBy(suite.validAccount)
	assert.NoError(suite.T(), err)
	
	// Test with future date - should fail
	futureSpec := NewAccountCreatedAfter(tomorrow)
	err = futureSpec.IsSatisfiedBy(suite.validAccount)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "created")
	assert.Contains(suite.T(), err.Error(), "before required date")
}

func (suite *AccountSpecificationTestSuite) TestAccountStatusIn() {
	spec := NewAccountStatusIn(
		v1.AccountStatus_ACCOUNT_STATUS_ACTIVE,
		v1.AccountStatus_ACCOUNT_STATUS_PENDING,
	)
	
	// Active account should pass
	err := spec.IsSatisfiedBy(suite.validAccount)
	assert.NoError(suite.T(), err)
	
	// Suspended account should fail
	err = spec.IsSatisfiedBy(suite.invalidAccount)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "not in allowed statuses")
}

// Test composite specifications

func (suite *AccountSpecificationTestSuite) TestValidAccountForBilling() {
	spec := NewValidAccountForBilling()
	
	// Valid account should pass all checks
	err := spec.IsSatisfiedBy(suite.validAccount)
	assert.NoError(suite.T(), err)
	
	// Invalid account should fail multiple checks
	err = spec.IsSatisfiedBy(suite.invalidAccount)
	assert.Error(suite.T(), err)
	
	// Error should contain multiple validation failures
	errorString := err.Error()
	assert.Contains(suite.T(), errorString, "account ID is required")
	assert.Contains(suite.T(), errorString, "userId is required")
	assert.Contains(suite.T(), errorString, "tariffId is required")
}

func (suite *AccountSpecificationTestSuite) TestActiveAccountWithValidTariff() {
	spec := NewActiveAccountWithValidTariff()
	
	// Valid account should pass
	err := spec.IsSatisfiedBy(suite.validAccount)
	assert.NoError(suite.T(), err)
	
	// Create account that's active but has no tariff
	activeNoTariff := &v1.Account{
		Id:       uuid.New().String(),
		UserId:   uuid.New().String(),
		TariffId: uuid.Nil.String(),
		Status:   v1.AccountStatus_ACCOUNT_STATUS_ACTIVE,
	}
	
	err = spec.IsSatisfiedBy(activeNoTariff)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "tariffId is required")
}

func (suite *AccountSpecificationTestSuite) TestAccountSuspendedOrExpired() {
	spec := NewAccountSuspendedOrExpired()
	
	// Suspended account should satisfy the spec
	suspendedAccount := &v1.Account{
		Id:     uuid.New().String(),
		Status: v1.AccountStatus_ACCOUNT_STATUS_SUSPENDED,
	}
	err := spec.IsSatisfiedBy(suspendedAccount)
	assert.NoError(suite.T(), err)
	
	// Expired account should satisfy the spec
	expiredAccount := &v1.Account{
		Id:        uuid.New().String(),
		Status:    v1.AccountStatus_ACCOUNT_STATUS_ACTIVE,
		ExpiresAt: timestamppb.New(time.Now().Add(-time.Hour)),
	}
	err = spec.IsSatisfiedBy(expiredAccount)
	assert.NoError(suite.T(), err)
	
	// Active, non-expired account should not satisfy the spec
	err = spec.IsSatisfiedBy(suite.validAccount)
	assert.Error(suite.T(), err)
}

// Test specification filtering

func (suite *AccountSpecificationTestSuite) TestFilterAccountsBySpecification() {
	accounts := []*v1.Account{
		suite.validAccount,
		suite.invalidAccount,
		{
			Id:       uuid.New().String(),
			UserId:   uuid.New().String(),
			TariffId: uuid.New().String(),
			Status:   v1.AccountStatus_ACCOUNT_STATUS_PENDING,
		},
	}
	
	spec := NewAccountActive()
	filtered, err := specification.Filter(accounts, spec)
	
	// Should have error because not all accounts are active
	assert.Error(suite.T(), err)
	// But should return the valid accounts that passed
	assert.Len(suite.T(), filtered, 1)
	assert.Equal(suite.T(), suite.validAccount.GetId(), filtered[0].GetId())
}

// Test AccountService

func (suite *AccountSpecificationTestSuite) TestAccountService_ValidateAccountForBilling() {
	service := NewAccountService()
	
	err := service.ValidateAccountForBilling(suite.validAccount)
	assert.NoError(suite.T(), err)
	
	err = service.ValidateAccountForBilling(suite.invalidAccount)
	assert.Error(suite.T(), err)
}

func (suite *AccountSpecificationTestSuite) TestAccountService_ValidateAccountForActivation() {
	service := NewAccountService()
	
	err := service.ValidateAccountForActivation(suite.validAccount)
	assert.NoError(suite.T(), err)
	
	err = service.ValidateAccountForActivation(suite.invalidAccount)
	assert.Error(suite.T(), err)
}

func (suite *AccountSpecificationTestSuite) TestAccountService_CanAccountBeDeactivated() {
	service := NewAccountService()
	
	canDeactivate := service.CanAccountBeDeactivated(suite.validAccount)
	assert.True(suite.T(), canDeactivate)
	
	canDeactivate = service.CanAccountBeDeactivated(suite.invalidAccount)
	assert.False(suite.T(), canDeactivate)
}

func (suite *AccountSpecificationTestSuite) TestAccountService_GetAccountsEligibleForUpgrade() {
	service := NewAccountService()
	
	accounts := []*v1.Account{suite.validAccount, suite.invalidAccount}
	eligible, err := service.GetAccountsEligibleForUpgrade(accounts)
	
	assert.Error(suite.T(), err) // Not all accounts are eligible
	assert.Len(suite.T(), eligible, 1)
	assert.Equal(suite.T(), suite.validAccount.GetId(), eligible[0].GetId())
}

// Test AccountSpecificationBuilder

func (suite *AccountSpecificationTestSuite) TestAccountSpecificationBuilder() {
	spec := NewAccountSpecificationBuilder().
		WithUserId().
		WithTariffId().
		WithActiveStatus().
		BuildAnd()
	
	err := spec.IsSatisfiedBy(suite.validAccount)
	assert.NoError(suite.T(), err)
	
	err = spec.IsSatisfiedBy(suite.invalidAccount)
	assert.Error(suite.T(), err)
}

func (suite *AccountSpecificationTestSuite) TestAccountSpecificationBuilder_EmptySpecs() {
	spec := NewAccountSpecificationBuilder().BuildAnd()
	
	// Should always pass with no specs
	err := spec.IsSatisfiedBy(suite.validAccount)
	assert.NoError(suite.T(), err)
	
	err = spec.IsSatisfiedBy(suite.invalidAccount)
	assert.NoError(suite.T(), err)
}

func (suite *AccountSpecificationTestSuite) TestAccountSpecificationBuilder_OrLogic() {
	// Account should pass if it has EITHER user ID OR tariff ID
	spec := NewAccountSpecificationBuilder().
		WithUserId().
		WithTariffId().
		BuildOr()
	
	// Account with both should pass
	err := spec.IsSatisfiedBy(suite.validAccount)
	assert.NoError(suite.T(), err)
	
	// Account with neither should fail
	err = spec.IsSatisfiedBy(suite.invalidAccount)
	assert.Error(suite.T(), err)
	
	// Account with only user ID should pass
	accountWithUserOnly := &v1.Account{
		Id:       uuid.New().String(),
		UserId:   uuid.New().String(),
		TariffId: uuid.Nil.String(),
		Status:   v1.AccountStatus_ACCOUNT_STATUS_ACTIVE,
	}
	err = spec.IsSatisfiedBy(accountWithUserOnly)
	assert.NoError(suite.T(), err)
}

// Test common specification combinations

func (suite *AccountSpecificationTestSuite) TestGetBasicValidationSpec() {
	spec := GetBasicValidationSpec()
	
	err := spec.IsSatisfiedBy(suite.validAccount)
	assert.NoError(suite.T(), err)
	
	err = spec.IsSatisfiedBy(suite.invalidAccount)
	assert.Error(suite.T(), err)
}

func (suite *AccountSpecificationTestSuite) TestGetFullValidationSpec() {
	spec := GetFullValidationSpec()
	
	err := spec.IsSatisfiedBy(suite.validAccount)
	assert.NoError(suite.T(), err)
	
	err = spec.IsSatisfiedBy(suite.invalidAccount)
	assert.Error(suite.T(), err)
}

func (suite *AccountSpecificationTestSuite) TestGetRenewalEligibilitySpec() {
	spec := GetRenewalEligibilitySpec()
	
	// Valid active account should be eligible
	err := spec.IsSatisfiedBy(suite.validAccount)
	assert.NoError(suite.T(), err)
	
	// Create expired but otherwise valid account
	expiredAccount := &v1.Account{
		Id:        uuid.New().String(),
		UserId:    uuid.New().String(),
		TariffId:  uuid.New().String(),
		Status:    v1.AccountStatus_ACCOUNT_STATUS_ACTIVE,
		ExpiresAt: timestamppb.New(time.Now().Add(-time.Hour)),
	}
	
	// Expired account should also be eligible for renewal
	err = spec.IsSatisfiedBy(expiredAccount)
	assert.NoError(suite.T(), err)
	
	// Invalid account should not be eligible
	err = spec.IsSatisfiedBy(suite.invalidAccount)
	assert.Error(suite.T(), err)
}

// Benchmark tests

func BenchmarkAccountSpecification_Simple(b *testing.B) {
	account := &v1.Account{
		Id:       uuid.New().String(),
		UserId:   uuid.New().String(),
		TariffId: uuid.New().String(),
		Status:   v1.AccountStatus_ACCOUNT_STATUS_ACTIVE,
	}
	spec := NewUserId()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = spec.IsSatisfiedBy(account)
	}
}

func BenchmarkAccountSpecification_Composite(b *testing.B) {
	account := &v1.Account{
		Id:       uuid.New().String(),
		UserId:   uuid.New().String(),
		TariffId: uuid.New().String(),
		Status:   v1.AccountStatus_ACCOUNT_STATUS_ACTIVE,
	}
	spec := NewValidAccountForBilling()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = spec.IsSatisfiedBy(account)
	}
}

func BenchmarkAccountSpecification_Filter(b *testing.B) {
	accounts := make([]*v1.Account, 1000)
	for i := range accounts {
		accounts[i] = &v1.Account{
			Id:       uuid.New().String(),
			UserId:   uuid.New().String(),
			TariffId: uuid.New().String(),
			Status:   v1.AccountStatus_ACCOUNT_STATUS_ACTIVE,
		}
	}
	spec := NewAccountActive()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = specification.Filter(accounts, spec)
	}
}

// Edge case tests

func TestAccountSpecification_NilAccount(t *testing.T) {
	spec := NewUserId()
	
	assert.Panics(t, func() {
		_ = spec.IsSatisfiedBy(nil)
	})
}

func TestAccountSpecification_EmptyUUID(t *testing.T) {
	account := &v1.Account{
		Id:       "",
		UserId:   "",
		TariffId: "",
	}
	
	spec := NewUserId()
	err := spec.IsSatisfiedBy(account)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUserIdRequired)
}

func TestAccountSpecification_ValidUUIDString(t *testing.T) {
	account := &v1.Account{
		Id:       uuid.New().String(),
		UserId:   uuid.New().String(),
		TariffId: uuid.New().String(),
	}
	
	userSpec := NewUserId()
	err := userSpec.IsSatisfiedBy(account)
	assert.NoError(t, err)
	
	tariffSpec := NewTariffId()
	err = tariffSpec.IsSatisfiedBy(account)
	assert.NoError(t, err)
	
	idSpec := NewAccountIdRequired()
	err = idSpec.IsSatisfiedBy(account)
	assert.NoError(t, err)
}