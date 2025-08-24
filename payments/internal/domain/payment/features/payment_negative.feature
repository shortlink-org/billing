Feature: Negative transitions

  Scenario: Double capture is rejected
    Given a payment "n1" is created for invoice "inv-n1"
      And the amount is "USD 10.00"
      And the payment kind is "ONE_TIME"
      And the capture mode is "IMMEDIATE"
      And I capture "USD 6.00"
    When I try to capture "USD 5.00"
    Then the operation must be rejected
      And the payment state must be "PAID"

  Scenario: Refund above captured is rejected
    Given a payment "n2" is created for invoice "inv-n2"
      And the amount is "USD 10.00"
      And the payment kind is "ONE_TIME"
      And the capture mode is "IMMEDIATE"
      And I capture "USD 7.00"
    When I try to refund "USD 8.00"
    Then the operation must be rejected
      And the payment state must be "PAID"

  Scenario: Cancel from PAID is rejected
    Given a payment "n3" is created for invoice "inv-n3"
      And the amount is "USD 10.00"
      And the payment kind is "ONE_TIME"
      And the capture mode is "IMMEDIATE"
      And I capture "USD 10.00"
    When I try to cancel the payment with reason "USER"
    Then the operation must be rejected
      And the payment state must be "PAID"
