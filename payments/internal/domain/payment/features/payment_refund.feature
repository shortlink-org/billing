Feature: Refund semantics

  Scenario: Partial refund keeps the payment in PAID
    Given a payment "r1" is created for invoice "inv-r1"
      And the amount is "USD 10.00"
      And the payment kind is "ONE_TIME"
      And the capture mode is "IMMEDIATE"
      And I capture "USD 10.00"
    When I refund "USD 3.00"
    Then the payment state must be "PAID"
      And the total refunded equals "USD 3.00"
      And full refund flag is "false"
      And the last uncommitted event is "PaymentRefunded"

  Scenario: Full refund moves the payment to REFUNDED
    Given a payment "r2" is created for invoice "inv-r2"
      And the amount is "USD 10.00"
      And the payment kind is "ONE_TIME"
      And the capture mode is "IMMEDIATE"
      And I capture "USD 10.00"
    When I refund "USD 10.00"
    Then the payment state must be "REFUNDED"
      And the total refunded equals "USD 10.00"
      And full refund flag is "true"
