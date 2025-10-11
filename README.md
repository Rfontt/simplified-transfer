![system business logic](./system_design/simplified_transfer_business_logic.png)

# Requirements

**functional:**

- To both users:
  - The full name, the users document (CPF/CNPJ), email and password are required fields

- To commons user:
  - A common user could transfer money to shopkeepers and others commons users
  - To transfer money, the user needs to have a positive balance in wallet.

- To shopkeepers user:
  - Just receive money transfers but they couldn't transfer it  

**Non - functional:**

- CPF/CNPJ and email must be unique fields
- Call this endpoint: https://util.devi.tools/api/v2/authorize to call a mock gateway authorization 
- The transfer operation needs be transactional, it means that if an error occur the money needs be refund in wallet
- User needs be notified when receive a transfer. The notification could be a sms or an email
  - The notify service could be unavailable, so when it happening use a fallback or put in a dlq queue to retry
- endpoint to transfer:

```
POST /transfer
Content-Type: application/json

{
  "value": 100.0,
  "payer": 4,
  "payee": 15
}
```

- The database will use CQRS
  - Commands: It will handle with a Postgres database
  - Queries: It will handle with a MongoDB database
- Consistency is not required, but the system must be available and partition-tolerant.

