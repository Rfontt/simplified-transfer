## Send an email to the user who received the transfer

```mermaid
flowchart LR
    TransferService -->|AccountTransferSucceeded| EventBus[(Kafka)]
    EventBus -->|Change account transaction status| AccountService
    EventBus --> |notify user who received the transfer| EmailService
```