Feature: Refunds (partial and full)

  Background:
    And the amount is "USD 100.00"
    And the payment kind is "ONE_TIME"
    And the capture mode is "IMMEDIATE"

  Scenario: Partial refund keeps state PAID
    Given a payment "44444444-4444-4444-4444-444444444444" is created for invoice "dddddddd-dddd-dddd-dddd-dddddddddddd"
    When I capture "USD 100.00"
    Then the payment state must be "PAID"

    When I refund "USD 40.00"
    Then the total refunded equals "USD 40.00"
    And the payment state must be "PAID"
    And full refund flag is "false"

  Scenario: Full refund moves state to REFUNDED
    Given a payment "55555555-5555-5555-5555-555555555555" is created for invoice "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"
    When I capture "USD 100.00"
    Then the payment state must be "PAID"

    When I refund "USD 100.00"
    Then the total refunded equals "USD 100.00"
    And the payment state must be "REFUNDED"
    And full refund flag is "true"

  Scenario: Refund fails but state remains PAID
    Given a payment "66666666-6666-6666-6666-666666666666" is created for invoice "ffffffff-ffff-ffff-ffff-ffffffffffff"
    When I capture "USD 100.00"
    And a refund attempt fails with reason "NETWORK_ERROR"
    Then the payment state must be "PAID"
