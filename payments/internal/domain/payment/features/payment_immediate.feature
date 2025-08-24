Feature: Immediate capture payments
  As a consumer of the Payments domain
  I want to complete a one-time payment with immediate capture
  So that funds are captured right after creation without authorization hold

  Scenario: Created -> Paid (immediate capture, full amount)
    Given a payment "pay-1" is created for invoice "inv-1"
      And the amount is "USD 10.00"
      And the payment kind is "ONE_TIME"
      And the capture mode is "IMMEDIATE"
    When I capture "USD 10.00"
    Then the payment state must be "PAID"
      And the captured total equals "USD 10.00"
      And the uncommitted events include, in order:
        | PaymentCreated |
        | PaymentPaid    |

  Scenario: Refund attempt fails does not change state
    Given a payment "pay-2" is created for invoice "inv-2"
      And the amount is "USD 9.00"
      And the payment kind is "ONE_TIME"
      And the capture mode is "IMMEDIATE"
      And I capture "USD 9.00"
    When a refund attempt fails with reason "NETWORK_ERROR"
    Then the payment state must still be "PAID"
      And the last uncommitted event is "PaymentRefundFailed"
