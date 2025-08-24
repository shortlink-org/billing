Feature: Happy paths

  Scenario: SCA → authorize → capture → full refund
    Given a payment "hp1" is created for invoice "inv-hp1"
      And the amount is "USD 10.00"
      And the payment kind is "ONE_TIME"
      And the capture mode is "IMMEDIATE"
    When I require SCA
      And I confirm authorization of "USD 10.00"
      And I capture "USD 10.00"
      And I refund "USD 10.00"
    Then the payment state must be "REFUNDED"
      And the total refunded equals "USD 10.00"
      And full refund flag is "true"

  Scenario: Immediate capture (allowed by FSM; domain Policy decides)
    Given a payment "hp2" is created for invoice "inv-hp2"
      And the amount is "USD 10.00"
      And the payment kind is "ONE_TIME"
      And the capture mode is "IMMEDIATE"
    When I capture "USD 10.00"
    Then the payment state must be "PAID"
      And the captured total equals "USD 10.00"
