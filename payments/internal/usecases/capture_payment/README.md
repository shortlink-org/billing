## Use Case: UC-3 Capture a previously authorized payment

### Description
This use case handles capturing (settling) a payment that was previously authorized. This is commonly used in scenarios where payment authorization happens before order fulfillment.

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

actor Merchant as merchant
participant "Payment Service" as payment_service
participant "Order Service" as order_service
participant "Database" as db
participant "Payment Gateway" as gateway
participant "Inventory Service" as inventory
participant "Event Bus" as events
participant "Accounting Service" as accounting

== Capture Payment ==
merchant -> payment_service ++: POST /payments/{id}/capture {amount, reason}
note right of payment_service #WAITING_COLOR: Processing payment capture

payment_service -> db ++: Find authorized payment
alt Payment found and authorized
    db --> payment_service --: SUCCESS_COLOR: Authorized payment retrieved
    payment_service -> order_service ++: Verify order fulfillment status
    alt Order fulfilled
        order_service --> payment_service --: SUCCESS_COLOR: Order ready for capture
        payment_service -> inventory ++: Confirm items shipped/delivered
        alt Items confirmed
            inventory --> payment_service --: SUCCESS_COLOR: Fulfillment confirmed
            payment_service -> gateway ++: Capture authorized payment
            note right of gateway #WAITING_COLOR: Processing capture with bank
            alt Capture successful
                gateway --> payment_service --: SUCCESS_COLOR: Payment captured
                payment_service -> db ++: Update payment status to captured
                alt Update successful
                    db --> payment_service --: SUCCESS_COLOR: Status updated
                    payment_service -> accounting ++: Record revenue
                    accounting --> payment_service --: SUCCESS_COLOR: Revenue recorded
                    payment_service -> events ++: Publish payment_captured event
                    events --> payment_service --: SUCCESS_COLOR: Event published
                    payment_service --> merchant --: SUCCESS_COLOR: 200 Payment Captured
                else Update failed
                    db --> payment_service --: ERROR_COLOR: Database error
                    payment_service -> gateway ++: Reverse capture
                    gateway --> payment_service --: WAITING_COLOR: Reversal initiated
                    payment_service --> merchant --: ERROR_COLOR: 500 Capture Error
                end
            else Capture failed
                gateway --> payment_service --: ERROR_COLOR: Capture declined
                payment_service -> db ++: Log capture failure
                db --> payment_service --: SUCCESS_COLOR: Failure logged
                payment_service -> events ++: Publish capture_failed event
                events --> payment_service --: ERROR_COLOR: Event published
                payment_service --> merchant --: ERROR_COLOR: 402 Capture Failed
            end
        else Items not fulfilled
            inventory --> payment_service --: ERROR_COLOR: Fulfillment incomplete
            payment_service --> merchant --: ERROR_COLOR: 400 Order Not Fulfilled
        end
    else Order not ready
        order_service --> payment_service --: ERROR_COLOR: Order not fulfilled
        payment_service --> merchant --: ERROR_COLOR: 400 Order Not Ready
    end
else Payment not found or wrong status
    db --> payment_service --: ERROR_COLOR: Payment not found/wrong status
    payment_service --> merchant --: ERROR_COLOR: 404 Payment Not Found
end

== Partial Capture ==
alt Partial capture requested
    merchant -> payment_service ++: POST /payments/{id}/capture {amount: partial}
    payment_service -> db ++: Validate partial amount <= authorized amount
    alt Valid partial amount
        db --> payment_service --: SUCCESS_COLOR: Partial amount valid
        payment_service -> gateway ++: Capture partial amount
        gateway --> payment_service --: SUCCESS_COLOR: Partial capture successful
        payment_service -> db ++: Update captured amount
        db --> payment_service --: SUCCESS_COLOR: Amount updated
        payment_service -> events ++: Publish partial_capture event
        events --> payment_service --: SUCCESS_COLOR: Event published
        payment_service --> merchant --: SUCCESS_COLOR: 200 Partial Capture Successful
    else Invalid partial amount
        db --> payment_service --: ERROR_COLOR: Amount exceeds authorization
        payment_service --> merchant --: ERROR_COLOR: 400 Invalid Amount
    end
end

== Authorization Expiry Check ==
participant "Expiry Monitor" as monitor
monitor -> payment_service ++: Check authorization expiry
payment_service -> db ++: Find expiring authorizations
db --> payment_service --: SUCCESS_COLOR: Expiring authorizations list
loop For each expiring authorization
    payment_service -> gateway ++: Check authorization status
    alt Authorization still valid
        gateway --> payment_service --: SUCCESS_COLOR: Still valid
        payment_service -> merchant ++: Notify of pending expiry
        merchant --> payment_service --: SUCCESS_COLOR: Notification received
    else Authorization expired
        gateway --> payment_service --: ERROR_COLOR: Authorization expired
        payment_service -> db ++: Update status to expired
        db --> payment_service --: SUCCESS_COLOR: Status updated
        payment_service -> events ++: Publish authorization_expired event
        events --> payment_service --: ERROR_COLOR: Event published
    end
end
payment_service --> monitor --: SUCCESS_COLOR: Expiry check completed

@enduml
```

### Error Scenarios
- **400 Bad Request**: Invalid capture amount or order not fulfilled
- **402 Payment Required**: Capture declined by payment gateway
- **404 Not Found**: Payment not found or not in authorized status
- **410 Gone**: Authorization has expired
- **500 Internal Error**: Database or service failures

### Success Scenarios
- **200 OK**: Payment successfully captured
- **Partial Capture**: Partial amount successfully captured with remaining authorization
- **Full Capture**: Complete authorized amount captured and settled