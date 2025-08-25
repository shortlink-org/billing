## Use Case: UC-2 Confirm a pending payment (SCA/3DS)

### Description
This use case handles the confirmation of a pending payment that requires Strong Customer Authentication (SCA) or 3D Secure (3DS) verification.

### Sequence Diagram

```plantuml
@startuml
!define SUCCESS_COLOR #90EE90
!define ERROR_COLOR #FFB6C1
!define WAITING_COLOR #FFFFE0

skinparam sequence {
    ArrowColor black
    LifeLineBorderColor black
    LifeLineBackgroundColor white
    ParticipantBorderColor black
    ParticipantBackgroundColor white
    ParticipantFontColor black
    ActorBorderColor black
    ActorBackgroundColor white
    ActorFontColor black
}

actor Customer as customer
participant "Payment Service" as payment_service
participant "Database" as db
participant "Payment Gateway" as gateway
participant "SCA/3DS Provider" as sca_provider
participant "Order Service" as order_service
participant "Event Bus" as events
participant "Notification Service" as notification

== Confirm Payment ==
customer -> payment_service ++: POST /payments/{id}/confirm {authentication_result}
note right of payment_service #WAITING_COLOR: Processing payment confirmation

payment_service -> db ++: Find pending payment
alt Payment found and pending
    db --> payment_service --: SUCCESS_COLOR: Payment record retrieved
    payment_service -> gateway ++: Validate authentication result
    note right of gateway #WAITING_COLOR: Validating SCA/3DS result
    alt Authentication valid
        gateway --> payment_service --: SUCCESS_COLOR: Authentication confirmed
        payment_service -> gateway ++: Process confirmed payment
        note right of gateway #WAITING_COLOR: Processing payment with bank
        alt Payment successful
            gateway --> payment_service --: SUCCESS_COLOR: Payment processed
            payment_service -> db ++: Update payment status to succeeded
            alt Update successful
                db --> payment_service --: SUCCESS_COLOR: Status updated
                payment_service -> order_service ++: Update order status to paid
                order_service --> payment_service --: SUCCESS_COLOR: Order updated
                payment_service -> events ++: Publish payment_confirmed event
                events --> payment_service --: SUCCESS_COLOR: Event published
                payment_service -> notification ++: Send payment confirmation
                notification --> payment_service --: SUCCESS_COLOR: Confirmation sent
                payment_service --> customer --: SUCCESS_COLOR: 200 Payment Confirmed
            else Update failed
                db --> payment_service --: ERROR_COLOR: Database error
                payment_service -> gateway ++: Initiate refund for processed payment
                gateway --> payment_service --: WAITING_COLOR: Refund initiated
                payment_service --> customer --: ERROR_COLOR: 500 Confirmation Error
            end
        else Payment failed
            gateway --> payment_service --: ERROR_COLOR: Payment declined
            payment_service -> db ++: Update payment status to failed
            db --> payment_service --: SUCCESS_COLOR: Status updated
            payment_service -> events ++: Publish payment_failed event
            events --> payment_service --: ERROR_COLOR: Event published
            payment_service --> customer --: ERROR_COLOR: 402 Payment Failed
        end
    else Authentication invalid
        gateway --> payment_service --: ERROR_COLOR: Authentication failed
        payment_service -> db ++: Update payment status to failed
        db --> payment_service --: SUCCESS_COLOR: Status updated
        payment_service -> events ++: Publish authentication_failed event
        events --> payment_service --: ERROR_COLOR: Event published
        payment_service --> customer --: ERROR_COLOR: 401 Authentication Failed
    end
else Payment not found or not pending
    db --> payment_service --: ERROR_COLOR: Payment not found/wrong status
    payment_service --> customer --: ERROR_COLOR: 404 Payment Not Found
end

== SCA Challenge Flow ==
alt SCA challenge required
    payment_service -> sca_provider ++: Initiate SCA challenge
    note right of sca_provider #WAITING_COLOR: Preparing SCA challenge
    sca_provider -> customer ++: Present SCA challenge (SMS/App/Biometric)
    note right of customer #WAITING_COLOR: Customer completing SCA
    customer --> sca_provider --: Complete SCA challenge
    sca_provider --> payment_service --: SUCCESS_COLOR: SCA completed
    payment_service -> gateway ++: Continue with SCA result
    gateway --> payment_service --: SUCCESS_COLOR: Payment authorized
end

== Timeout Handling ==
participant "Timeout Handler" as timeout
timeout -> payment_service ++: Payment confirmation timeout
payment_service -> db ++: Find expired pending payments
db --> payment_service --: SUCCESS_COLOR: Expired payments list
loop For each expired payment
    payment_service -> db ++: Update status to timeout
    db --> payment_service --: SUCCESS_COLOR: Status updated
    payment_service -> events ++: Publish payment_timeout event
    events --> payment_service --: WAITING_COLOR: Event published
    payment_service -> notification ++: Notify customer of timeout
    notification --> payment_service --: SUCCESS_COLOR: Notification sent
end
payment_service --> timeout --: SUCCESS_COLOR: Timeout handling completed

@enduml
```

### Error Scenarios
- **401 Unauthorized**: SCA/3DS authentication failed
- **402 Payment Required**: Payment declined by bank/card issuer
- **404 Not Found**: Payment not found or not in pending status
- **408 Request Timeout**: Payment confirmation timeout exceeded
- **500 Internal Error**: Database or service failures

### Success Scenarios
- **200 OK**: Payment successfully confirmed and processed
- **SCA Challenge**: Additional authentication step completed successfully