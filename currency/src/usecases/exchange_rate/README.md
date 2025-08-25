# Use Case: Fetch and Store Exchange Rates

> **Note:**
> This use case outlines the process of fetching exchange rates from multiple providers and storing them in the system.

## Overview

This use case describes how the **Currency Service** fetches real-time exchange rates from multiple external providers, such as **Bloomberg** and **Yahoo**, and stores the retrieved rates in both a **Cache** and a **Database** for future use. The system ensures data accuracy and availability by implementing retries and fallback mechanisms when necessary.

## Flow

1. **Client Requests Exchange Rate**:

    - The **Client** sends a request to the **Currency Service** to fetch the exchange rate for a specific currency pair (e.g., USD/EUR).

2. **Check Cache for Rate**:

    - The service first checks the **Cache** to see if the exchange rate is already available.
    - If a **Cache Hit** occurs, the service retrieves the rate from the cache and returns it to the client.
    - If a **Cache Miss** occurs, the service proceeds to fetch the rate from external providers.

3. **Fetch Exchange Rates from Providers**:

    - The **Currency Service** requests exchange rates from multiple providers (e.g., **Bloomberg** and **Yahoo**).
    - The service implements retries with exponential backoff in case of temporary issues with the providers.
    - If one provider fails, the service falls back to another provider.

4. **Store Rates in Cache and Database**:

    - Once a valid exchange rate is retrieved, the service stores the rate in the **Cache** for quick future access.
    - The rate is also stored in the **Database** for long-term storage and historical reference.

5. **Return the Exchange Rate**:

    - The service returns the fetched exchange rate to the client.

## Components Involved

- **Currency Service**: Core service responsible for fetching exchange rates and storing them in both cache and database.
- **Cache Store**: Temporary storage for frequently accessed exchange rates to reduce external API calls.
- **Database**: Permanent storage for exchange rates, used for long-term access and historical queries.
- **External Providers**: Services like **Bloomberg** and **Yahoo** that provide real-time exchange rate data.

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
participant "Currency Service" as service
participant "Cache Store" as cache
participant "Bloomberg API" as bloomberg
participant "Yahoo API" as yahoo
participant "Database" as db

== Fetch Exchange Rate ==
client -> service ++: GET /rates/USD/EUR
note right of service #WAITING_COLOR: Processing rate request

service -> cache ++: Check cache for USD/EUR
alt Cache Hit
    cache --> service --: SUCCESS_COLOR: Return cached rate (1.12)
    service --> client --: SUCCESS_COLOR: 200 OK + Rate from cache
else Cache Miss
    cache --> service --: ERROR_COLOR: Cache miss
    service -> bloomberg ++: Fetch rate from Bloomberg
    note right of bloomberg #WAITING_COLOR: Fetching from Bloomberg
    alt Bloomberg Success
        bloomberg --> service --: SUCCESS_COLOR: Return rate (1.12)
        service -> cache ++: Store rate in cache
        cache --> service --: SUCCESS_COLOR: Rate cached
        service -> db ++: Store rate in database
        db --> service --: SUCCESS_COLOR: Rate stored
        service --> client --: SUCCESS_COLOR: 200 OK + Fetched rate
    else Bloomberg Fails
        bloomberg --> service --: ERROR_COLOR: Bloomberg API error
        service -> yahoo ++: Fetch rate from Yahoo (fallback)
        note right of yahoo #WAITING_COLOR: Fetching from Yahoo
        alt Yahoo Success
            yahoo --> service --: SUCCESS_COLOR: Return rate (1.12)
            service -> cache ++: Store rate in cache
            cache --> service --: SUCCESS_COLOR: Rate cached
            service -> db ++: Store rate in database
            db --> service --: SUCCESS_COLOR: Rate stored
            service --> client --: SUCCESS_COLOR: 200 OK + Fetched rate
        else Yahoo Fails
            yahoo --> service --: ERROR_COLOR: Yahoo API error
            service --> client --: ERROR_COLOR: 503 Service Unavailable
        end
    end
end

== Error Handling Scenarios ==
note over client, service
Color Coding:
- Green (SUCCESS_COLOR): Successful operations, rate found
- Red (ERROR_COLOR): API failures, cache miss, service errors
- Yellow (WAITING_COLOR): Processing states, fetching data
end note

@enduml
```

## Error Handling

### 1. Cache Unavailable

- **Scenario**: The cache store is unavailable when the system attempts to retrieve or store exchange rates.
- **Handling**:
    - The service skips the cache step and directly fetches the exchange rate from the external providers.
    - The service logs the cache failure and proceeds with the operation.

### 2. Provider Unavailable

- **Scenario**: One of the external exchange rate providers (e.g., **Bloomberg**) is unavailable.
- **Handling**:
    - The service retries the request with exponential backoff.
    - If the provider is still unavailable, the service falls back to another provider (e.g., **Yahoo**).
    - The error is logged for further investigation.

### 3. Rate Not Found

- **Scenario**: The requested exchange rate for the specified currency pair is not found in any provider.
- **Handling**:
    - The service returns an error response (`404 Not Found`) to the client.
    - The error is logged for auditing purposes, and the client is advised to check the currency pair.

### 4. Database Unavailable

- **Scenario**: The database is unavailable when the service tries to store the exchange rate.
- **Handling**:
    - The service still returns the fetched rate to the client, but logs the database error.
    - The exchange rate is only stored in the cache for temporary access, and a retry mechanism may be triggered later to store the rate in the database.
