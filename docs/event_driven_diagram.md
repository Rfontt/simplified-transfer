## User and Account Creation Flow

```mermaid
sequenceDiagram
    participant User
    participant API
    participant UserService
    participant EventBus
    participant AccountService

    User->>API: Sign up
    API->>UserService: CreateUserCommand
    UserService->>UserService: Validate & persist User
    UserService-->>EventBus: UserCreated event
    EventBus-->>AccountService: UserCreated event
    AccountService->>AccountService: Create Account
```
