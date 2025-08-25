## Use Case: UC-01 Works with Account

**Functional Requirements:**

1. CRUD account (Create, Read, List, Update, Delete)

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
participant "Account Service" as service
participant "Database" as db
participant "Validator" as validator

== Create Account ==
client -> service ++: POST /accounts
note right of service #WAITING_COLOR: Processing request
service -> validator ++: Validate account data
alt Valid data
    validator --> service --: SUCCESS_COLOR: Valid
    service -> db ++: Create account
    alt Success
        db --> service --: SUCCESS_COLOR: Account created
        service --> client --: SUCCESS_COLOR: 201 Created
    else Database error
        db --> service --: ERROR_COLOR: DB Error
        service --> client --: ERROR_COLOR: 500 Internal Error
    end
else Invalid data
    validator --> service --: ERROR_COLOR: Validation failed
    service --> client --: ERROR_COLOR: 400 Bad Request
end

== Read Account ==
client -> service ++: GET /accounts/{id}
note right of service #WAITING_COLOR: Fetching account
service -> db ++: Find account by ID
alt Found
    db --> service --: SUCCESS_COLOR: Account data
    service --> client --: SUCCESS_COLOR: 200 OK + Account
else Not found
    db --> service --: ERROR_COLOR: Not found
    service --> client --: ERROR_COLOR: 404 Not Found
end

== Update Account ==
client -> service ++: PUT /accounts/{id}
note right of service #WAITING_COLOR: Updating account
service -> validator ++: Validate update data
alt Valid data
    validator --> service --: SUCCESS_COLOR: Valid
    service -> db ++: Update account
    alt Success
        db --> service --: SUCCESS_COLOR: Account updated
        service --> client --: SUCCESS_COLOR: 200 OK
    else Not found
        db --> service --: ERROR_COLOR: Account not found
        service --> client --: ERROR_COLOR: 404 Not Found
    end
else Invalid data
    validator --> service --: ERROR_COLOR: Validation failed
    service --> client --: ERROR_COLOR: 400 Bad Request
end

== Delete Account ==
client -> service ++: DELETE /accounts/{id}
note right of service #WAITING_COLOR: Deleting account
service -> db ++: Delete account
alt Success
    db --> service --: SUCCESS_COLOR: Account deleted
    service --> client --: SUCCESS_COLOR: 204 No Content
else Not found
    db --> service --: ERROR_COLOR: Account not found
    service --> client --: ERROR_COLOR: 404 Not Found
end

@enduml
```
