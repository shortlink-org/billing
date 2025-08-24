### Flow of a Payment

```plantuml
@startuml
left to right direction
skinparam linetype ortho
skinparam shadowing false
skinparam roundcorner 16
skinparam dpi 120
skinparam nodesep 110
skinparam ranksep 110
skinparam ArrowThickness 1.2
skinparam ArrowFontSize 12
skinparam state {
  Padding 12
  BackgroundColor<<good>>  #A5D6A7
  BackgroundColor<<bad>>   #EF9A9A
  BackgroundColor<<wait>>  #FFF59D
  BorderColor Black
  FontColor Black
}

[*] --> Created

' --- In Progress (yellow) ----------------------------------------------------
state "In Progress" as IP {
  state "Created\n(intent created)" as Created <<wait>>
  state "WaitingForConfirmation\n(3DS/SCA, optional)" as WFC <<wait>>
  state "Authorized\n(funds reserved)" as Authorized <<wait>>

  Created -[#2E7D32,bold]-> WFC       : SCA step (optional)
  WFC    -[#2E7D32,bold]-> Authorized : confirm

  ' optional: skip SCA
  Created -[#888]-> Authorized : authorize
}

' --- Success (green, on the right) ------------------------------------------
state "Success" as S {
  [*] --> Paid
  state "Paid\n(captured)" as Paid <<good>>
  state "Refunded\n(full/partial)" as Refunded <<good>>
  Paid -[#2E7D32,bold]-> Refunded : refund
}

' --- Problem (red, below) ----------------------------------------------------
state "Problem" as P {
  [*] --> Router
  state Router <<choice>>
  state "Canceled" as Canceled <<bad>>
  state "Failed"   as Failed   <<bad>>
  Router -[#C62828]-> Canceled : cancel / void
  Router -[#C62828]-> Failed   : fail / reverse / expire
}

' --- Aggregated exits from "In Progress" ------------------------------------
IP -[#2E7D32,bold]-> S : capture / confirm→capture / immediate
IP -[#C62828]-> P      : cancel / void / fail / reverse / expire

' --- layout nudges (use hidden arrows; no direction with [hidden])
IP -[hidden]-> S
S  -[hidden]-> P

Canceled --> [*]
Failed   --> [*]
Refunded --> [*]
@enduml
```

### Refund semantics

- Partial refunds do **not** change the state: the payment remains **PAID**.
- A **full** refund moves the payment to **REFUNDED** (terminal).
- Additional refunds after full refund are not allowed.