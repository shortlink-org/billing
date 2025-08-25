## Use Case: Credit Card

> [!NOTE]
>
> This use case handles billing operations related to credit cards,
> including CRUD operations and validation through the Luhn algorithm.

### Use Cases

#### 1. Create Credit Card

- Add a new credit card entry with card details such as number, expiration date, etc.
- Applies Luhn validation to ensure the card number is valid.

#### 2. Read Credit Card

- Retrieves details of a specific credit card based on its unique identifier.

#### 3. Update Credit Card

- Modifies existing credit card's information (e.g., expiration date).

#### 4. Delete Credit Card

- Removes a credit card from the system.

#### 5. Validate Credit Card

- Run a check using the Luhn algorithm to verify that a credit card number is valid before processing any transactions. [(Luhn algorithm on Wikipedia)](https://en.wikipedia.org/wiki/Luhn_algorithm)

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
participant "Credit Card Service" as service
participant "Luhn Validator" as validator
participant "Database" as db
participant "Encryption Service" as encryption
participant "Antifraud Service" as antifraud

== Add Credit Card ==
customer -> service ++: POST /cards
note right of service #WAITING_COLOR: Processing new card
service -> validator ++: Validate card number (Luhn)
alt Luhn check passed
    validator --> service --: SUCCESS_COLOR: Card number valid
    service -> antifraud ++: Check for suspicious activity
    alt Fraud check passed
        antifraud --> service --: SUCCESS_COLOR: No fraud detected
        service -> encryption ++: Encrypt card data
        encryption --> service --: SUCCESS_COLOR: Data encrypted
        service -> db ++: Store encrypted card
        alt Success
            db --> service --: SUCCESS_COLOR: Card stored
            service --> customer --: SUCCESS_COLOR: 201 Card Added
        else Database error
            db --> service --: ERROR_COLOR: Storage failed
            service --> customer --: ERROR_COLOR: 500 Internal Error
        end
    else Fraud detected
        antifraud --> service --: ERROR_COLOR: Suspicious activity
        service --> customer --: ERROR_COLOR: 403 Forbidden
    end
else Luhn check failed
    validator --> service --: ERROR_COLOR: Invalid card number
    service --> customer --: ERROR_COLOR: 400 Invalid Card Number
end

== Read Credit Card ==
customer -> service ++: GET /cards/{id}
note right of service #WAITING_COLOR: Fetching card details
service -> db ++: Find card by ID
alt Found
    db --> service --: SUCCESS_COLOR: Encrypted card data
    service -> encryption ++: Decrypt card data
    encryption --> service --: SUCCESS_COLOR: Decrypted data
    service --> customer --: SUCCESS_COLOR: 200 OK + Card Details (masked)
else Not found
    db --> service --: ERROR_COLOR: Card not found
    service --> customer --: ERROR_COLOR: 404 Not Found
end

== Update Credit Card ==
customer -> service ++: PUT /cards/{id}
note right of service #WAITING_COLOR: Updating card
service -> db ++: Find existing card
alt Card found
    db --> service --: SUCCESS_COLOR: Card found
    service -> validator ++: Validate new expiry date
    alt Valid expiry
        validator --> service --: SUCCESS_COLOR: Expiry valid
        service -> encryption ++: Encrypt updated data
        encryption --> service --: SUCCESS_COLOR: Data encrypted
        service -> db ++: Update card
        alt Success
            db --> service --: SUCCESS_COLOR: Card updated
            service --> customer --: SUCCESS_COLOR: 200 OK
        else Database error
            db --> service --: ERROR_COLOR: Update failed
            service --> customer --: ERROR_COLOR: 500 Internal Error
        end
    else Invalid expiry
        validator --> service --: ERROR_COLOR: Invalid expiry date
        service --> customer --: ERROR_COLOR: 400 Invalid Expiry
    end
else Card not found
    db --> service --: ERROR_COLOR: Card not found
    service --> customer --: ERROR_COLOR: 404 Not Found
end

== Delete Credit Card ==
customer -> service ++: DELETE /cards/{id}
note right of service #WAITING_COLOR: Removing card
service -> db ++: Check if card has pending transactions
alt No pending transactions
    db --> service --: SUCCESS_COLOR: Safe to delete
    service -> db ++: Delete card
    db --> service --: SUCCESS_COLOR: Card deleted
    service --> customer --: SUCCESS_COLOR: 204 No Content
else Has pending transactions
    db --> service --: ERROR_COLOR: Pending transactions exist
    service --> customer --: ERROR_COLOR: 400 Cannot Delete
end

== Validate Credit Card (Luhn Algorithm) ==
customer -> service ++: POST /cards/validate
note right of service #WAITING_COLOR: Validating card
service -> validator ++: Apply Luhn algorithm
alt Luhn check passed
    validator --> service --: SUCCESS_COLOR: Card number valid
    service --> customer --: SUCCESS_COLOR: 200 Card Valid
else Luhn check failed
    validator --> service --: ERROR_COLOR: Invalid card number
    service --> customer --: ERROR_COLOR: 400 Invalid Card Number
end

@enduml
```
