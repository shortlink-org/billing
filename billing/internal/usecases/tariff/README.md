## Use Case: UC-04 Works with tariff

**Functional Requirements:**

1. CRUD tariff (Create, Read, List, Update, Delete)

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

actor Admin as admin
actor Customer as customer
participant "Tariff Service" as service
participant "Database" as db
participant "Pricing Engine" as pricing
participant "Notification Service" as notification

== Create Tariff (Admin) ==
admin -> service ++: POST /tariffs
note right of service #WAITING_COLOR: Creating new tariff
service -> pricing ++: Validate pricing rules
alt Valid pricing
    pricing --> service --: SUCCESS_COLOR: Pricing valid
    service -> db ++: Create tariff
    alt Success
        db --> service --: SUCCESS_COLOR: Tariff created
        service -> notification ++: Notify stakeholders
        notification --> service --: SUCCESS_COLOR: Notifications sent
        service --> admin --: SUCCESS_COLOR: 201 Tariff Created
    else Database error
        db --> service --: ERROR_COLOR: DB Error
        service --> admin --: ERROR_COLOR: 500 Internal Error
    end
else Invalid pricing
    pricing --> service --: ERROR_COLOR: Invalid pricing
    service --> admin --: ERROR_COLOR: 400 Bad Request
end

== Read Tariff ==
customer -> service ++: GET /tariffs/{id}
note right of service #WAITING_COLOR: Fetching tariff
service -> db ++: Find tariff by ID
alt Found and active
    db --> service --: SUCCESS_COLOR: Tariff data
    service --> customer --: SUCCESS_COLOR: 200 OK + Tariff
else Not found or inactive
    db --> service --: ERROR_COLOR: Not found
    service --> customer --: ERROR_COLOR: 404 Not Found
end

== List Available Tariffs ==
customer -> service ++: GET /tariffs
note right of service #WAITING_COLOR: Fetching available tariffs
service -> db ++: Get active tariffs
alt Found
    db --> service --: SUCCESS_COLOR: Active tariffs
    service --> customer --: SUCCESS_COLOR: 200 OK + Tariffs List
else No tariffs
    db --> service --: SUCCESS_COLOR: Empty list
    service --> customer --: SUCCESS_COLOR: 200 OK + Empty List
end

== Update Tariff (Admin) ==
admin -> service ++: PUT /tariffs/{id}
note right of service #WAITING_COLOR: Updating tariff
service -> pricing ++: Validate new pricing
alt Valid pricing
    pricing --> service --: SUCCESS_COLOR: Pricing valid
    service -> db ++: Update tariff
    alt Success
        db --> service --: SUCCESS_COLOR: Tariff updated
        service -> notification ++: Notify customers of changes
        notification --> service --: SUCCESS_COLOR: Notifications sent
        service --> admin --: SUCCESS_COLOR: 200 OK
    else Not found
        db --> service --: ERROR_COLOR: Tariff not found
        service --> admin --: ERROR_COLOR: 404 Not Found
    end
else Invalid pricing
    pricing --> service --: ERROR_COLOR: Invalid pricing
    service --> admin --: ERROR_COLOR: 400 Bad Request
end

== Deactivate Tariff (Admin) ==
admin -> service ++: DELETE /tariffs/{id}
note right of service #WAITING_COLOR: Deactivating tariff
service -> db ++: Check if tariff has active subscriptions
alt No active subscriptions
    db --> service --: SUCCESS_COLOR: Can deactivate
    service -> db ++: Deactivate tariff
    db --> service --: SUCCESS_COLOR: Tariff deactivated
    service --> admin --: SUCCESS_COLOR: 200 Tariff Deactivated
else Has active subscriptions
    db --> service --: ERROR_COLOR: Active subscriptions exist
    service --> admin --: ERROR_COLOR: 400 Cannot Deactivate
end

@enduml
```
