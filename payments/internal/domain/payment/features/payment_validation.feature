Feature: Domain validations on amounts and limits

  Scenario: Cannot authorize above amount
    Given a payment "pay-8" is created for invoice "inv-8"
      And the amount is "USD 10.00"
      And the payment kind is "ONE_TIME"
      And the capture mode is "MANUAL"
    When I try to authorize "USD 11.00"
    Then the operation must be rejected
      And the payment state must be "CREATED"

  Scenario: Cannot capture more than authorized
    Given a payment "pay-9" is created for invoice "inv-9"
      And the amount is "USD 10.00"
      And the payment kind is "ONE_TIME"
      And the capture mode is "MANUAL"
      And I authorize "USD 6.00"
    When I try to capture "USD 7.00"
    Then the operation must be rejected
      And the payment state must be "AUTHORIZED"

  Scenario: USD minor-unit arithmetic respects cap
    Given a payment "pay-10" is created for invoice "inv-10"
      And the amount is "USD 0.09"
      And the payment kind is "ONE_TIME"
      And the capture mode is "MANUAL"
    When I authorize "USD 0.05"
      And I authorize "USD 0.04"
    Then the authorized total equals "USD 0.09"
    When I try to authorize "USD 0.01"
    Then the operation must be rejected
    # Optional: show that nanos beyond scale are rejected by validation, too:
    When I try to authorize "USD 0.005000000"
    Then the operation must be rejected
