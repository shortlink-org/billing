## Use Case: UC-4 Refund a payment (full or partial)

### Description
This use case handles refunding a previously captured payment, either in full or partially. It includes validation, inventory updates, and accounting adjustments.

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
actor Customer as customer
participant "Payment Service" as payment_service
participant "Order Service" as order_service
participant "Database" as db
participant "Payment Gateway" as gateway
participant "Inventory Service" as inventory
participant "Event Bus" as events
participant "Accounting Service" as accounting
participant "Notification Service" as notification

== Initiate Refund ==
merchant -> payment_service ++: POST /payments/{id}/refund {amount, reason}
note right of payment_service #WAITING_COLOR: Processing refund request

payment_service -> db ++: Find captured payment
alt Payment found and captured
    db --> payment_service --: SUCCESS_COLOR: Payment record retrieved
    payment_service -> db ++: Validate refund amount
    alt Valid refund amount
        db --> payment_service --: SUCCESS_COLOR: Refund amount valid
        payment_service -> order_service ++: Check order return status
        alt Return approved
            order_service --> payment_service --: SUCCESS_COLOR: Return authorized
            payment_service -> inventory ++: Update inventory if needed
            alt Inventory updated
                inventory --> payment_service --: SUCCESS_COLOR: Inventory adjusted
                payment_service -> gateway ++: Process refund
                note right of gateway #WAITING_COLOR: Processing refund with bank
                alt Refund successful
                    gateway --> payment_service --: SUCCESS_COLOR: Refund processed
                    payment_service -> db ++: Create refund record
                    alt Record created
                        db --> payment_service --: SUCCESS_COLOR: Refund recorded
                        payment_service -> accounting ++: Adjust revenue
                        accounting --> payment_service --: SUCCESS_COLOR: Revenue adjusted
                        payment_service -> events ++: Publish refund_processed event
                        events --> payment_service --: SUCCESS_COLOR: Event published
                        payment_service -> notification ++: Notify customer
                        notification --> payment_service --: SUCCESS_COLOR: Customer notified
                        payment_service --> merchant --: SUCCESS_COLOR: 200 Refund Processed
                    else Record creation failed
                        db --> payment_service --: ERROR_COLOR: Database error
                        payment_service -> gateway ++: Reverse refund
                        gateway --> payment_service --: WAITING_COLOR: Reversal initiated
                        payment_service --> merchant --: ERROR_COLOR: 500 Refund Error
                    end
                else Refund failed
                    gateway --> payment_service --: ERROR_COLOR: Refund declined
                    payment_service -> inventory ++: Revert inventory changes
                    inventory --> payment_service --: SUCCESS_COLOR: Inventory reverted
                    payment_service --> merchant --: ERROR_COLOR: 402 Refund Failed
                end
            else Inventory update failed
                inventory --> payment_service --: ERROR_COLOR: Inventory error
                payment_service --> merchant --: ERROR_COLOR: 500 Inventory Error
            end
        else Return not approved
            order_service --> payment_service --: ERROR_COLOR: Return not authorized
            payment_service --> merchant --: ERROR_COLOR: 400 Return Not Approved
        end
    else Invalid refund amount
        db --> payment_service --: ERROR_COLOR: Amount exceeds refundable
        payment_service --> merchant --: ERROR_COLOR: 400 Invalid Amount
    end
else Payment not found or not refundable
    db --> payment_service --: ERROR_COLOR: Payment not found/not refundable
    payment_service --> merchant --: ERROR_COLOR: 404 Payment Not Found
end

== Customer-Initiated Refund ==
customer -> payment_service ++: POST /payments/{id}/request-refund {reason}
note right of payment_service #WAITING_COLOR: Processing customer refund request

payment_service -> db ++: Find customer's payment
alt Payment found and belongs to customer
    db --> payment_service --: SUCCESS_COLOR: Payment verified
    payment_service -> order_service ++: Check return policy
    alt Within return window
        order_service --> payment_service --: SUCCESS_COLOR: Return eligible
        payment_service -> db ++: Create refund request
        db --> payment_service --: SUCCESS_COLOR: Request created
        payment_service -> merchant ++: Notify merchant of refund request
        merchant --> payment_service --: SUCCESS_COLOR: Merchant notified
        payment_service --> customer --: SUCCESS_COLOR: 202 Refund Request Submitted
    else Outside return window
        order_service --> payment_service --: ERROR_COLOR: Return window expired
        payment_service --> customer --: ERROR_COLOR: 400 Return Window Expired
    end
else Payment not found or unauthorized
    db --> payment_service --: ERROR_COLOR: Payment not found/unauthorized
    payment_service --> customer --: ERROR_COLOR: 404 Payment Not Found
end

== Partial Refund ==
alt Partial refund requested
    merchant -> payment_service ++: POST /payments/{id}/refund {amount: partial}
    payment_service -> db ++: Check remaining refundable amount
    alt Sufficient refundable amount
        db --> payment_service --: SUCCESS_COLOR: Partial refund valid
        payment_service -> gateway ++: Process partial refund
        gateway --> payment_service --: SUCCESS_COLOR: Partial refund processed
        payment_service -> db ++: Update refunded amounts
        db --> payment_service --: SUCCESS_COLOR: Amounts updated
        payment_service -> events ++: Publish partial_refund event
        events --> payment_service --: SUCCESS_COLOR: Event published
        payment_service --> merchant --: SUCCESS_COLOR: 200 Partial Refund Processed
    else Insufficient refundable amount
        db --> payment_service --: ERROR_COLOR: Amount exceeds refundable
        payment_service --> merchant --: ERROR_COLOR: 400 Insufficient Refundable Amount
    end
end

== Automatic Refund (Chargeback) ==
participant "Chargeback Handler" as chargeback
gateway -> chargeback ++: Chargeback notification
chargeback -> payment_service ++: Process automatic refund
payment_service -> db ++: Find payment for chargeback
db --> payment_service --: SUCCESS_COLOR: Payment found
payment_service -> db ++: Create chargeback refund
db --> payment_service --: SUCCESS_COLOR: Chargeback recorded
payment_service -> accounting ++: Adjust for chargeback
accounting --> payment_service --: SUCCESS_COLOR: Adjustments made
payment_service -> events ++: Publish chargeback_processed event
events --> payment_service --: ERROR_COLOR: Event published
payment_service -> notification ++: Notify merchant of chargeback
notification --> payment_service --: ERROR_COLOR: Merchant notified
payment_service --> chargeback --: ERROR_COLOR: Chargeback processed
chargeback --> gateway --: SUCCESS_COLOR: Chargeback handled

@enduml
```

### Error Scenarios
- **400 Bad Request**: Invalid refund amount or return not approved
- **402 Payment Required**: Refund declined by payment gateway
- **404 Not Found**: Payment not found or not refundable
- **410 Gone**: Refund window has expired
- **500 Internal Error**: Database or service failures

### Success Scenarios
- **200 OK**: Refund successfully processed
- **202 Accepted**: Refund request submitted for approval
- **Partial Refund**: Partial amount successfully refunded
- **Chargeback Handling**: Automatic refund processed for chargeback