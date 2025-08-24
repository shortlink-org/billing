Feature: Domain policy

  Scenario: Immediate capture is forbidden in MANUAL mode
    Given a payment "p1" is created for invoice "inv-p1"
      And the amount is "USD 10.00"
      And the payment kind is "ONE_TIME"
      And the capture mode is "MANUAL"
    When I try to capture "USD 10.00"
    Then the operation must be rejected
      And the payment state must be "CREATED"

  Scenario: Immediate capture is allowed in IMMEDIATE mode
    Given a payment "p2" is created for invoice "inv-p2"
      And the amount is "USD 10.00"
      And the payment kind is "ONE_TIME"
      And the capture mode is "IMMEDIATE"
    When I capture "USD 10.00"
    Then the payment state must be "PAID"
