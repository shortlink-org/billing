## UC-5: Works with a subscription

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

actor Customer as customer
actor Admin as admin
participant "Subscription Service" as service
participant "Database" as db
participant "Stripe API" as stripe
participant "Billing Service" as billing
participant "Notification Service" as notification

== Create Subscription ==
customer -> service ++: POST /subscriptions
note right of service #WAITING_COLOR: Creating subscription
service -> db ++: Validate customer and tariff
alt Valid data
    db --> service --: SUCCESS_COLOR: Validation passed
    service -> stripe ++: Create Stripe subscription
    note right of stripe #WAITING_COLOR: Processing with Stripe
    alt Stripe success
        stripe --> service --: SUCCESS_COLOR: Subscription created
        service -> db ++: Save subscription
        alt Success
            db --> service --: SUCCESS_COLOR: Subscription saved
            service -> notification ++: Send welcome email
            notification --> service --: SUCCESS_COLOR: Email sent
            service --> customer --: SUCCESS_COLOR: 201 Subscription Created
        else Database error
            db --> service --: ERROR_COLOR: Save failed
            service -> stripe ++: Cancel Stripe subscription
            stripe --> service --: Cleanup completed
            service --> customer --: ERROR_COLOR: 500 Internal Error
        end
    else Stripe failure
        stripe --> service --: ERROR_COLOR: Payment failed
        service --> customer --: ERROR_COLOR: 402 Payment Required
    end
else Invalid data
    db --> service --: ERROR_COLOR: Validation failed
    service --> customer --: ERROR_COLOR: 400 Bad Request
end

== Cancel Subscription ==
customer -> service ++: DELETE /subscriptions/{id}
note right of service #WAITING_COLOR: Cancelling subscription
service -> db ++: Find subscription
alt Subscription found and active
    db --> service --: SUCCESS_COLOR: Subscription found
    service -> stripe ++: Cancel Stripe subscription
    alt Stripe success
        stripe --> service --: SUCCESS_COLOR: Cancelled in Stripe
        service -> billing ++: Calculate prorated refund
        billing --> service --: SUCCESS_COLOR: Refund calculated
        service -> stripe ++: Process refund
        stripe --> service --: SUCCESS_COLOR: Refund processed
        service -> db ++: Update subscription status
        db --> service --: SUCCESS_COLOR: Status updated
        service -> notification ++: Send cancellation email
        notification --> service --: SUCCESS_COLOR: Email sent
        service --> customer --: SUCCESS_COLOR: 200 Subscription Cancelled
    else Stripe failure
        stripe --> service --: ERROR_COLOR: Cancellation failed
        service --> customer --: ERROR_COLOR: 500 Cancellation Failed
    end
else Subscription not found or inactive
    db --> service --: ERROR_COLOR: Not found
    service --> customer --: ERROR_COLOR: 404 Not Found
end

== View Subscription ==
customer -> service ++: GET /subscriptions/{id}
note right of service #WAITING_COLOR: Fetching subscription
service -> db ++: Find subscription
alt Found
    db --> service --: SUCCESS_COLOR: Subscription data
    service --> customer --: SUCCESS_COLOR: 200 OK + Subscription
else Not found
    db --> service --: ERROR_COLOR: Not found
    service --> customer --: ERROR_COLOR: 404 Not Found
end

== Admin Manage Subscription ==
admin -> service ++: PUT /subscriptions/{id}/admin
note right of service #WAITING_COLOR: Admin managing subscription
service -> db ++: Update subscription
alt Success
    db --> service --: SUCCESS_COLOR: Subscription updated
    service -> notification ++: Notify customer of changes
    notification --> service --: SUCCESS_COLOR: Customer notified
    service --> admin --: SUCCESS_COLOR: 200 OK
else Not found
    db --> service --: ERROR_COLOR: Subscription not found
    service --> admin --: ERROR_COLOR: 404 Not Found
end

@enduml
```
