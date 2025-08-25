## Antifraud

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
participant "Payment Service" as payment
participant "Antifraud Service" as antifraud
participant "Risk Engine" as risk
participant "Machine Learning" as ml
participant "Database" as db
participant "Notification Service" as notification

== Fraud Detection Process ==
payment -> antifraud ++: Validate transaction {amount, card, customer}
note right of antifraud #WAITING_COLOR: Analyzing transaction

antifraud -> risk ++: Evaluate risk factors
note right of risk #WAITING_COLOR: Calculating risk score
risk -> ml ++: Get ML fraud prediction
alt High risk indicators
    ml --> risk --: ERROR_COLOR: High fraud probability (0.85)
    risk --> antifraud --: ERROR_COLOR: Risk score: HIGH
    antifraud -> db ++: Log suspicious transaction
    db --> antifraud --: SUCCESS_COLOR: Transaction logged
    antifraud -> notification ++: Alert fraud team
    notification --> antifraud --: SUCCESS_COLOR: Alert sent
    antifraud --> payment --: ERROR_COLOR: TRANSACTION_BLOCKED
else Medium risk indicators
    ml --> risk --: WAITING_COLOR: Medium fraud probability (0.45)
    risk --> antifraud --: WAITING_COLOR: Risk score: MEDIUM
    antifraud -> db ++: Log transaction for review
    db --> antifraud --: SUCCESS_COLOR: Transaction logged
    antifraud -> notification ++: Queue for manual review
    notification --> antifraud --: WAITING_COLOR: Queued for review
    antifraud --> payment --: WAITING_COLOR: MANUAL_REVIEW_REQUIRED
else Low risk indicators
    ml --> risk --: SUCCESS_COLOR: Low fraud probability (0.15)
    risk --> antifraud --: SUCCESS_COLOR: Risk score: LOW
    antifraud -> db ++: Log approved transaction
    db --> antifraud --: SUCCESS_COLOR: Transaction logged
    antifraud --> payment --: SUCCESS_COLOR: TRANSACTION_APPROVED
end

== Transaction Status Updates ==
alt Transaction Blocked
    payment -> customer: Transaction declined due to security
else Manual Review Required
    payment -> customer: Transaction pending review
    note right of customer #WAITING_COLOR: Customer waits for approval
else Transaction Approved
    payment -> customer: Transaction processed successfully
end

== Manual Review Process (Background) ==
participant "Fraud Analyst" as analyst
analyst -> antifraud ++: Review flagged transaction
antifraud -> db ++: Get transaction details
db --> antifraud --: SUCCESS_COLOR: Transaction data
alt Analyst approves
    antifraud -> db ++: Update status to approved
    db --> antifraud --: SUCCESS_COLOR: Status updated
    antifraud -> payment ++: Process approved transaction
    payment --> antifraud --: SUCCESS_COLOR: Transaction processed
    antifraud --> analyst --: SUCCESS_COLOR: Transaction approved
else Analyst rejects
    antifraud -> db ++: Update status to rejected
    db --> antifraud --: SUCCESS_COLOR: Status updated
    antifraud -> notification ++: Notify customer of rejection
    notification --> antifraud --: SUCCESS_COLOR: Customer notified
    antifraud --> analyst --: ERROR_COLOR: Transaction rejected
end

@enduml
```
