## Use Case: UC-03 Works with payment

**Functional Requirements:**

1. CRUD payment (Create, Read, List, Update, Delete)

## Sequence Diagram

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

actor Client as client
participant "Payment Service" as service
participant "Database" as db
participant "Payment Gateway" as gateway
participant "Order Service" as order
participant "Antifraud Service" as antifraud

== Create Payment ==
client -> service ++: POST /payments
note right of service #WAITING_COLOR: Processing payment
service -> order ++: Validate order exists
alt Order valid
    order --> service --: SUCCESS_COLOR: Order validated
    service -> antifraud ++: Check for fraud
    alt Fraud check passed
        antifraud --> service --: SUCCESS_COLOR: Payment approved
        service -> gateway ++: Process payment
        note right of gateway #WAITING_COLOR: Processing with bank
        alt Payment successful
            gateway --> service --: SUCCESS_COLOR: Payment processed
            service -> db ++: Store payment record
            alt Success
                db --> service --: SUCCESS_COLOR: Payment stored
                service -> order ++: Update order status
                order --> service --: SUCCESS_COLOR: Order updated
                service --> client --: SUCCESS_COLOR: 201 Payment Created
            else Database error
                db --> service --: ERROR_COLOR: Storage failed
                service --> client --: ERROR_COLOR: 500 Internal Error
            end
        else Payment failed
            gateway --> service --: ERROR_COLOR: Payment declined
            service --> client --: ERROR_COLOR: 402 Payment Required
        end
    else Fraud detected
        antifraud --> service --: ERROR_COLOR: Fraud detected
        service --> client --: ERROR_COLOR: 403 Forbidden
    end
else Invalid order
    order --> service --: ERROR_COLOR: Order not found
    service --> client --: ERROR_COLOR: 400 Invalid Order
end

== Read Payment ==
client -> service ++: GET /payments/{id}
note right of service #WAITING_COLOR: Fetching payment
service -> db ++: Find payment by ID
alt Found
    db --> service --: SUCCESS_COLOR: Payment data
    service --> client --: SUCCESS_COLOR: 200 OK + Payment
else Not found
    db --> service --: ERROR_COLOR: Not found
    service --> client --: ERROR_COLOR: 404 Not Found
end

== Refund Payment ==
client -> service ++: POST /payments/{id}/refund
note right of service #WAITING_COLOR: Processing refund
service -> db ++: Find payment
alt Payment found and refundable
    db --> service --: SUCCESS_COLOR: Payment found
    service -> gateway ++: Process refund
    note right of gateway #WAITING_COLOR: Processing refund
    alt Refund successful
        gateway --> service --: SUCCESS_COLOR: Refund processed
        service -> db ++: Update payment status
        db --> service --: SUCCESS_COLOR: Payment updated
        service --> client --: SUCCESS_COLOR: 200 Refund Processed
    else Refund failed
        gateway --> service --: ERROR_COLOR: Refund failed
        service --> client --: ERROR_COLOR: 400 Refund Failed
    end
else Payment not refundable
    db --> service --: ERROR_COLOR: Cannot refund
    service --> client --: ERROR_COLOR: 400 Cannot Refund
end

@enduml
```
