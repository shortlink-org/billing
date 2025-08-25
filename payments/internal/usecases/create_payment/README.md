## Use Case: UC-1 Create a payment for an invoice/order

### Description
This use case handles the creation of a new payment for an invoice or order. It includes validation, fraud checking, and payment gateway integration.

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
participant "Order Service" as order_service
participant "Validation Service" as validator
participant "Antifraud Service" as antifraud
participant "Payment Gateway" as gateway
participant "Database" as db
participant "Event Bus" as events

== Create Payment ==
customer -> payment_service ++: POST /payments {order_id, payment_method, amount}
note right of payment_service #WAITING_COLOR: Processing payment creation

payment_service -> validator ++: Validate payment request
alt Valid request
    validator --> payment_service --: SUCCESS_COLOR: Validation passed
    payment_service -> order_service ++: Verify order exists and amount
    alt Order valid
        order_service --> payment_service --: SUCCESS_COLOR: Order verified
        payment_service -> antifraud ++: Check for fraud
        note right of antifraud #WAITING_COLOR: Analyzing transaction risk
        alt Fraud check passed
            antifraud --> payment_service --: SUCCESS_COLOR: Low risk score
            payment_service -> gateway ++: Create payment intent
            note right of gateway #WAITING_COLOR: Creating payment with provider
            alt Payment intent created
                gateway --> payment_service --: SUCCESS_COLOR: Payment intent ID
                payment_service -> db ++: Store payment record
                alt Storage successful
                    db --> payment_service --: SUCCESS_COLOR: Payment stored
                    payment_service -> events ++: Publish payment_created event
                    events --> payment_service --: SUCCESS_COLOR: Event published
                    payment_service --> customer --: SUCCESS_COLOR: 201 Payment Created
                else Storage failed
                    db --> payment_service --: ERROR_COLOR: Database error
                    payment_service -> gateway ++: Cancel payment intent
                    gateway --> payment_service --: Cleanup completed
                    payment_service --> customer --: ERROR_COLOR: 500 Storage Error
                end
            else Payment intent failed
                gateway --> payment_service --: ERROR_COLOR: Gateway error
                payment_service --> customer --: ERROR_COLOR: 502 Gateway Error
            end
        else Fraud detected
            antifraud --> payment_service --: ERROR_COLOR: High risk score
            payment_service -> db ++: Log fraud attempt
            db --> payment_service --: SUCCESS_COLOR: Fraud logged
            payment_service --> customer --: ERROR_COLOR: 403 Transaction Blocked
        end
    else Order invalid
        order_service --> payment_service --: ERROR_COLOR: Order not found/invalid
        payment_service --> customer --: ERROR_COLOR: 400 Invalid Order
    end
else Invalid request
    validator --> payment_service --: ERROR_COLOR: Validation failed
    payment_service --> customer --: ERROR_COLOR: 400 Bad Request
end

== 3D Secure Flow (if required) ==
alt Payment requires 3DS
    gateway -> customer ++: Redirect to 3DS authentication
    note right of customer #WAITING_COLOR: Customer authenticating
    customer --> gateway --: Complete 3DS challenge
    gateway -> payment_service ++: 3DS authentication result
    alt 3DS successful
        payment_service -> db ++: Update payment status
        db --> payment_service --: SUCCESS_COLOR: Status updated
        payment_service -> events ++: Publish payment_authenticated event
        events --> payment_service --: SUCCESS_COLOR: Event published
        payment_service --> gateway --: SUCCESS_COLOR: Continue processing
    else 3DS failed
        payment_service -> db ++: Update payment status to failed
        db --> payment_service --: SUCCESS_COLOR: Status updated
        payment_service -> events ++: Publish payment_failed event
        events --> payment_service --: ERROR_COLOR: Event published
        payment_service --> gateway --: ERROR_COLOR: Payment failed
    end
end

@enduml
```

### Error Scenarios
- **400 Bad Request**: Invalid payment data or order ID
- **403 Forbidden**: Transaction blocked due to fraud detection
- **404 Not Found**: Order not found
- **500 Internal Error**: Database or internal service failures
- **502 Bad Gateway**: Payment gateway communication errors

### Success Scenarios
- **201 Created**: Payment successfully created and stored
- **Payment requires additional authentication**: 3DS flow initiated