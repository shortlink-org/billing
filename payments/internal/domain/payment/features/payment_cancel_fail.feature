Feature: Problem exits (cancel/fail)

  Background:
    And the amount is "USD 5.00"
    And the payment kind is "ONE_TIME"

  Scenario: Created can be canceled
    Given a payment "77777777-7777-7777-7777-777777777777" is created for invoice "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
    And the capture mode is "IMMEDIATE"
    When I cancel the payment with reason "SYSTEM"
    Then the payment state must be "CANCELED"

  Scenario: Authorization expired -> FAILED
    Given a payment "88888888-8888-8888-8888-888888888888" is created for invoice "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
    And the capture mode is "MANUAL"
    When I fail the payment with reason "AUTH_EXPIRED"
    Then the payment state must be "FAILED"

  Scenario: Declined -> FAILED
    Given a payment "99999999-9999-9999-9999-999999999999" is created for invoice "cccccccc-cccc-cccc-cccc-cccccccccccc"
    And the capture mode is "IMMEDIATE"
    When I fail the payment with reason "DECLINED"
    Then the payment state must be "FAILED"
