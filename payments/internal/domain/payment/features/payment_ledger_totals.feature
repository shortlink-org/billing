Feature: Ledger totals are accumulated deterministically

  Background:
    And the amount is "USD 50.00"
    And the payment kind is "ONE_TIME"
    And the capture mode is "MANUAL"

  Scenario: Multiple incremental authorizations and captures
    Given a payment "dddddddd-eeee-ffff-0000-111111111111" is created for invoice "abababab-abab-abab-abab-abababababab"

    When I authorize "USD 20.00"
    And I authorize "USD 10.00"
    Then the authorized total equals "USD 30.00"

    When I capture "USD 5.00"
    And I capture "USD 25.00"
    Then the payment state must be "PAID"
    And the captured total equals "USD 30.00"

    When I refund "USD 10.00"
    Then the total refunded equals "USD 10.00"
    And the payment state must be "PAID"

    When I refund "USD 20.00"
    Then the total refunded equals "USD 30.00"
    And the payment state must be "REFUNDED"
