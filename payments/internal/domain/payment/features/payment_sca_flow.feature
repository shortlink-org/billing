Feature: SCA/3DS confirmation flow

  Background:
    And the amount is "USD 25.00"
    And the payment kind is "ONE_TIME"
    And the capture mode is "MANUAL"

  Scenario: Require SCA then confirm with incremental authorization
    Given a payment "33333333-3333-3333-3333-333333333333" is created for invoice "cccccccc-cccc-cccc-cccc-cccccccccccc"
    When I require SCA
    Then the payment state must be "WAITING_FOR_CONFIRMATION"

    When I confirm authorization of "USD 25.00"
    Then the payment state must be "AUTHORIZED"
    And the authorized total equals "USD 25.00"

    When I capture "USD 25.00"
    Then the payment state must be "PAID"
