Feature: Manual capture with optional SCA
  To support 3DS/SCA and delayed capture we authorize first and capture later

  Scenario: Created -> WaitingForConfirmation -> Authorized -> Paid
    Given a payment "pay-3" is created for invoice "inv-3"
      And the amount is "USD 10.00"
      And the payment kind is "ONE_TIME"
      And the capture mode is "MANUAL"
    When I require SCA
      And I confirm authorization of "USD 10.00"
      And I capture "USD 10.00"
    Then the payment state must be "PAID"
      And the authorized total equals "USD 10.00"
      And the captured total equals "USD 10.00"
      And the uncommitted events include, in order:
        | PaymentCreated                 |
        | PaymentWaitingForConfirmation  |
        | PaymentAuthorized              |
        | PaymentPaid                    |
