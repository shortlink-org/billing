Feature: Invalid transitions should be rejected

  Background:
    And the amount is "USD 10.00"
    And the payment kind is "ONE_TIME"

  Scenario: Capture more than authorized (MANUAL)
    Given a payment "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee" is created for invoice "11111111-2222-3333-4444-555555555555"
    And the capture mode is "MANUAL"
    When I authorize "USD 6.00"
    And I try to capture "USD 7.00"
    Then the operation must be rejected
    And the payment state must still be "AUTHORIZED"

  Scenario: Authorize more than amount
    Given a payment "bbbbbbbb-cccc-dddd-eeee-ffffffffffff" is created for invoice "66666666-7777-8888-9999-000000000000"
    And the capture mode is "MANUAL"
    When I authorize "USD 8.00"
    And I try to authorize "USD 3.00"
    Then the operation must be rejected
    And the authorized total equals "USD 8.00"

  Scenario: Refund without capture is rejected
    Given a payment "cccccccc-dddd-eeee-ffff-000000000000" is created for invoice "12121212-3434-5656-7878-909090909090"
    And the capture mode is "IMMEDIATE"
    When I try to refund "USD 1.00"
    Then the operation must be rejected
    And the payment state must still be "CREATED"
