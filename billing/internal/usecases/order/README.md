## Use Case: UC-02 Works with order

**Functional Requirements:**

1. CRUD order (Create, Read, List, Update, Delete)

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
participant "Order Service" as service
participant "Database" as db
participant "Inventory Service" as inventory
participant "Account Service" as account

== Create Order ==
client -> service ++: POST /orders
note right of service #WAITING_COLOR: Processing order
service -> account ++: Validate customer account
alt Account valid
    account --> service --: SUCCESS_COLOR: Account validated
    service -> inventory ++: Check product availability
    alt Products available
        inventory --> service --: SUCCESS_COLOR: Products available
        service -> db ++: Create order
        alt Success
            db --> service --: SUCCESS_COLOR: Order created
            service -> inventory ++: Reserve products
            inventory --> service --: SUCCESS_COLOR: Products reserved
            service --> client --: SUCCESS_COLOR: 201 Order Created
        else Database error
            db --> service --: ERROR_COLOR: DB Error
            service --> client --: ERROR_COLOR: 500 Internal Error
        end
    else Insufficient inventory
        inventory --> service --: ERROR_COLOR: Insufficient stock
        service --> client --: ERROR_COLOR: 400 Out of Stock
    end
else Invalid account
    account --> service --: ERROR_COLOR: Invalid account
    service --> client --: ERROR_COLOR: 400 Invalid Customer
end

== Read Order ==
client -> service ++: GET /orders/{id}
note right of service #WAITING_COLOR: Fetching order
service -> db ++: Find order by ID
alt Found
    db --> service --: SUCCESS_COLOR: Order data
    service --> client --: SUCCESS_COLOR: 200 OK + Order
else Not found
    db --> service --: ERROR_COLOR: Not found
    service --> client --: ERROR_COLOR: 404 Not Found
end

== Update Order Status ==
client -> service ++: PUT /orders/{id}/status
note right of service #WAITING_COLOR: Updating order status
service -> db ++: Update order status
alt Success
    db --> service --: SUCCESS_COLOR: Order updated
    service --> client --: SUCCESS_COLOR: 200 OK
else Not found
    db --> service --: ERROR_COLOR: Order not found
    service --> client --: ERROR_COLOR: 404 Not Found
end

== Cancel Order ==
client -> service ++: DELETE /orders/{id}
note right of service #WAITING_COLOR: Cancelling order
service -> db ++: Find order
alt Order found and cancellable
    db --> service --: SUCCESS_COLOR: Order found
    service -> inventory ++: Release reserved products
    inventory --> service --: SUCCESS_COLOR: Products released
    service -> db ++: Update order status to cancelled
    db --> service --: SUCCESS_COLOR: Order cancelled
    service --> client --: SUCCESS_COLOR: 200 Order Cancelled
else Order not found or not cancellable
    db --> service --: ERROR_COLOR: Cannot cancel
    service --> client --: ERROR_COLOR: 400 Cannot Cancel
end

@enduml
```
