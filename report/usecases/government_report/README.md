## Reports for Government

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

actor "Government Official" as gov
participant "Report Service" as service
participant "Data Aggregator" as aggregator
participant "Billing Database" as billing_db
participant "Transaction Database" as trans_db
participant "Tax Calculator" as tax_calc
participant "Document Generator" as doc_gen
participant "Secure Storage" as storage

== Generate Government Report ==
gov -> service ++: Request government report {period, type}
note right of service #WAITING_COLOR: Processing government report request

service -> aggregator ++: Aggregate financial data for period
note right of aggregator #WAITING_COLOR: Collecting data from multiple sources

aggregator -> billing_db ++: Get billing transactions
alt Data available
    billing_db --> aggregator --: SUCCESS_COLOR: Billing data retrieved
    aggregator -> trans_db ++: Get payment transactions
    alt Data available
        trans_db --> aggregator --: SUCCESS_COLOR: Transaction data retrieved
        aggregator -> tax_calc ++: Calculate tax obligations
        alt Calculation successful
            tax_calc --> aggregator --: SUCCESS_COLOR: Tax calculations completed
            aggregator --> service --: SUCCESS_COLOR: Aggregated data ready
            service -> doc_gen ++: Generate report document
            note right of doc_gen #WAITING_COLOR: Generating PDF/Excel report
            alt Document generation successful
                doc_gen --> service --: SUCCESS_COLOR: Report document generated
                service -> storage ++: Store report securely
                alt Storage successful
                    storage --> service --: SUCCESS_COLOR: Report stored
                    service --> gov --: SUCCESS_COLOR: 200 OK + Report ready for download
                else Storage failed
                    storage --> service --: ERROR_COLOR: Storage failed
                    service --> gov --: ERROR_COLOR: 500 Storage Error
                end
            else Document generation failed
                doc_gen --> service --: ERROR_COLOR: Document generation failed
                service --> gov --: ERROR_COLOR: 500 Generation Error
            end
        else Tax calculation failed
            tax_calc --> aggregator --: ERROR_COLOR: Tax calculation error
            aggregator --> service --: ERROR_COLOR: Calculation failed
            service --> gov --: ERROR_COLOR: 500 Calculation Error
        end
    else Transaction data unavailable
        trans_db --> aggregator --: ERROR_COLOR: No transaction data
        aggregator --> service --: ERROR_COLOR: Incomplete data
        service --> gov --: ERROR_COLOR: 404 Data Not Available
    end
else Billing data unavailable
    billing_db --> aggregator --: ERROR_COLOR: No billing data
    aggregator --> service --: ERROR_COLOR: No data available
    service --> gov --: ERROR_COLOR: 404 Data Not Available
end

== Download Report ==
gov -> service ++: Download report {report_id}
note right of service #WAITING_COLOR: Retrieving report
service -> storage ++: Get stored report
alt Report found
    storage --> service --: SUCCESS_COLOR: Report file retrieved
    service --> gov --: SUCCESS_COLOR: 200 OK + Report file
else Report not found
    storage --> service --: ERROR_COLOR: Report not found
    service --> gov --: ERROR_COLOR: 404 Report Not Found
end

== Schedule Periodic Reports ==
participant "Scheduler" as scheduler
scheduler -> service ++: Trigger scheduled government report
note right of service #WAITING_COLOR: Auto-generating periodic report
service -> aggregator ++: Aggregate data for reporting period
aggregator --> service --: SUCCESS_COLOR: Data aggregated
service -> doc_gen ++: Generate scheduled report
doc_gen --> service --: SUCCESS_COLOR: Report generated
service -> storage ++: Store scheduled report
storage --> service --: SUCCESS_COLOR: Report stored
service -> gov ++: Notify report availability
gov --> service --: SUCCESS_COLOR: Notification received
service --> scheduler --: SUCCESS_COLOR: Scheduled report completed

@enduml
```
