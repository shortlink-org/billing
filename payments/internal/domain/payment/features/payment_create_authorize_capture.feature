Feature: Create, authorize and capture payments

  Background:
    And the amount is "USD 100.00"
    And the payment kind is "ONE_TIME"

  Scenario: Created (IMMEDIATE) can capture without prior auth
    Given a payment "11111111-1111-1111-1111-111111111111" is created for invoice "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
    And the capture mode is "IMMEDIATE"
    When I capture "USD 100.00"
    Then the payment state must be "PAID"
    And the captured total equals "USD 100.00"
    And the uncommitted events include, in order:
      | PaymentCreated |
      | PaymentPaid    |

  Scenario: Created (MANUAL) must authorize first, then capture
    Given a payment "22222222-2222-2222-2222-222222222222" is created for invoice "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
    And the capture mode is "MANUAL"
    When I try to capture "USD 100.00"
    Then the operation must be rejected
    And the payment state must still be "CREATED"

    When I authorize "USD 60.00"
    And I authorize "USD 40.00"
    Then the authorized total equals "USD 100.00"
    And the payment state must be "AUTHORIZED"
    And the uncommitted events include, in order:
      | PaymentCreated   |
      | PaymentAuthorized|
      | PaymentAuthorized|

    When I capture "USD 70.00"
    And I capture "USD 30.00"
    Then the payment state must be "PAID"
    And the captured total equals "USD 100.00"
