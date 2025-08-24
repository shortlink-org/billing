Feature: Problem exits (cancel/fail)

  Scenario: Created can be canceled
    Given a payment "c1" is created for invoice "inv-c1"
      And the amount is "USD 5.00"
      And the payment kind is "ONE_TIME"
      And the capture mode is "IMMEDIATE"
    When I cancel the payment with reason "SYSTEM"
    Then the payment state must be "CANCELED"

  Scenario: Waiting can fail
    Given a payment "f1" is created for invoice "inv-f1"
      And the amount is "USD 5.00"
      And the payment kind is "ONE_TIME"
      And the capture mode is "IMMEDIATE"
      And I require SCA
    When I fail the payment with reason "DECLINED"
    Then the payment state must be "FAILED"
