# Use Case: Scheduled Updates of Exchange Rate Data via Cron Job

> **Note**  
> This use case covers the scheduled update process where a cron job runs at predefined intervals to refresh exchange rate data from Bloomberg and Yahoo APIs, ensuring that the data is up to date.

## Flow

1. **Scheduled Trigger**:  
   A cron job is set up to run at regular intervals (e.g., hourly). This job triggers the **Currency Service** to refresh exchange rate data from external providers. The purpose is to keep the system's cache and database populated with the latest rates, reducing the need to make API calls for every user request.

2. **Fetch Exchange Rates**:  
   The **Currency Service** executes the `load_exchange_rates` use case to fetch exchange rates from both **Bloomberg** and **Yahoo** APIs.  
   By using this [Load Exchange Rate Data from Subscriptions](#use-case-load-exchange-rate-data-from-subscriptions) use case, it ensures that the system maintains an up-to-date set of exchange rates.

3. **Handle API Responses**:
   - **Success**: If the API calls to **Bloomberg** and **Yahoo** are successful, the service retrieves the updated exchange rate data.
   - **Failure and Fallback**: If one API fails or is unavailable, the service will attempt to use the alternative provider. Additionally, it implements retries with exponential backoff to handle transient failures and avoid overwhelming the external providers.

4. **Store Rates in Cache and Database**:  
   Once the service successfully retrieves the updated rates, it stores them in the **Cache** for quick access and in the **Rate Database** for long-term storage and auditing.

5. **Logging and Monitoring**:  
   Throughout the process, the service logs the activities and any errors encountered, such as API failures or rate limit exceedances. This logging is crucial for monitoring the health of the system and for debugging issues when they occur.

6. **Completion**:  
   Once the data has been fetched, processed, and stored, the cron job execution is considered complete. The service waits until the next scheduled interval to perform another update.

## Sequence Diagram

### Cron Job Interaction Diagram

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

actor "Cron Job" as cron
participant "Currency Service" as service
participant "Cache Store" as cache
participant "Bloomberg API" as bloomberg
participant "Yahoo API" as yahoo
participant "Rate Database" as db
participant "Monitoring" as monitor

== Scheduled Rate Refresh ==
cron -> service ++: Trigger scheduled rate refresh
note right of service #WAITING_COLOR: Starting scheduled refresh

service -> bloomberg ++: Fetch latest rates from Bloomberg
note right of bloomberg #WAITING_COLOR: Fetching from Bloomberg
alt Bloomberg Success
    bloomberg --> service --: SUCCESS_COLOR: Return updated rates
    service -> yahoo ++: Fetch latest rates from Yahoo
    note right of yahoo #WAITING_COLOR: Fetching from Yahoo
    alt Yahoo Success
        yahoo --> service --: SUCCESS_COLOR: Return updated rates
        service -> cache ++: Update cache with new rates
        alt Cache Update Success
            cache --> service --: SUCCESS_COLOR: Cache updated
            service -> db ++: Update database with new rates
            alt Database Update Success
                db --> service --: SUCCESS_COLOR: Database updated
                service -> monitor ++: Log successful refresh
                monitor --> service --: SUCCESS_COLOR: Logged
                service --> cron --: SUCCESS_COLOR: Refresh completed
            else Database Update Failed
                db --> service --: ERROR_COLOR: Database update failed
                service -> monitor ++: Log database error
                monitor --> service --: ERROR_COLOR: Error logged
                service --> cron --: ERROR_COLOR: Partial failure
            end
        else Cache Update Failed
            cache --> service --: ERROR_COLOR: Cache update failed
            service -> db ++: Update database with new rates
            db --> service --: SUCCESS_COLOR: Database updated
            service -> monitor ++: Log cache warning
            monitor --> service --: WAITING_COLOR: Warning logged
            service --> cron --: WAITING_COLOR: Completed with warnings
        end
    else Yahoo Failed
        yahoo --> service --: ERROR_COLOR: Yahoo API error
        service -> cache ++: Update cache with Bloomberg rates only
        cache --> service --: SUCCESS_COLOR: Cache updated
        service -> db ++: Update database with Bloomberg rates
        db --> service --: SUCCESS_COLOR: Database updated
        service -> monitor ++: Log Yahoo failure
        monitor --> service --: ERROR_COLOR: Error logged
        service --> cron --: WAITING_COLOR: Partial success
    end
else Bloomberg Failed
    bloomberg --> service --: ERROR_COLOR: Bloomberg API error
    service -> yahoo ++: Fetch rates from Yahoo (fallback)
    alt Yahoo Success
        yahoo --> service --: SUCCESS_COLOR: Return updated rates
        service -> cache ++: Update cache with Yahoo rates
        cache --> service --: SUCCESS_COLOR: Cache updated
        service -> db ++: Update database with Yahoo rates
        db --> service --: SUCCESS_COLOR: Database updated
        service -> monitor ++: Log Bloomberg failure
        monitor --> service --: ERROR_COLOR: Error logged
        service --> cron --: WAITING_COLOR: Completed with fallback
    else Both APIs Failed
        yahoo --> service --: ERROR_COLOR: Yahoo API error
        service -> monitor ++: Log critical failure
        monitor --> service --: ERROR_COLOR: Critical error logged
        service --> cron --: ERROR_COLOR: Refresh failed
    end
end

== Color Legend ==
note over cron, service
Color Coding:
- Green (SUCCESS_COLOR): Successful operations
- Red (ERROR_COLOR): API failures, critical errors
- Yellow (WAITING_COLOR): Processing states, partial success, warnings
end note

@enduml
```

## Error Handling

### 1. API Failure

- **Scenario**: One of the external APIs (Bloomberg or Yahoo) fails to return data or is unavailable.
- **Handling**:
   - The service retries the request with exponential backoff.
   - If the API continues to fail, the system falls back to the alternative provider.
   - The failure is logged, and an alert may be triggered depending on the severity and duration of the outage.

### 2. Rate Limit Exceeded

- **Scenario**: The number of requests sent to the external API exceeds its rate limits.
- **Handling**:
   - The service tracks API usage and implements rate limiting to prevent further requests from exceeding the limits.
   - Retries are scheduled based on the API's rate limit reset time.
   - Errors are logged for monitoring and further analysis.

### 3. Cache or Database Unavailable

- **Scenario**: The cache or database is unavailable when the service attempts to store the updated exchange rates.
- **Handling**:
   - If the **Cache** is unavailable, the service will skip updating the cache and log the error.
   - If the **Rate Database** is unavailable, the system attempts to retry at a later time.
   - All failures are logged, and alerts may be triggered if necessary.

### 4. Invalid or Inconsistent Data

- **Scenario**: The data retrieved from the external provider is invalid or inconsistent.
- **Handling**:
   - The service discards invalid data and logs the issue.
   - It attempts to fetch the data again during the next scheduled cron job.
   - The system will continue serving cached or previously stored data to clients until valid data is retrieved.
