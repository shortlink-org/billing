Feature: Emitted events order for common flows

  Background:
    And the amount is "USD 12.00"
    And the payment kind is "ONE_TIME"

  Scenario: IMMEDIATE capture flow
    Given a payment "eeeeeeee-ffff-0000-1111-222222222222" is created for invoice "cdcdcdcd-cdcd-cdcd-cdcd-cdcdcdcdcdcd"
    And the capture mode is "IMMEDIATE"
    When I capture "USD 12.00"
    Then the uncommitted events include, in order:
      | PaymentCreated |
      | PaymentPaid    |

  Scenario: MANUAL with SCA
    Given a payment "ffffffff-0000-1111-2222-333333333333" is created for invoice "dededede-dede-dede-dede-dededededede"
    And the capture mode is "MANUAL"
    When I require SCA
    And I confirm authorization of "USD 12.00"
    And I capture "USD 12.00"
    Then the uncommitted events include, in order:
      | PaymentCreated                 |
      | PaymentWaitingForConfirmation  |
      | PaymentAuthorized              |
      | PaymentPaid                    |
