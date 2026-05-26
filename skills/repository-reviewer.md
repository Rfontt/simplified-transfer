---
name: repository-reviewer
description: Expert reviewer for Domain-Driven Design repositories in Go. Validates repository design against Evans and Vernon's DDD principles. Use when creating, reviewing, or validating Go repositories in your DDD project. Provides assertive feedback on the 12 DDD best practices, identifies anti-patterns, suggests proper method names using Ubiquitous Language, and determines if something should be a Repository or Query Object.
---

# DDD Go Repository Reviewer

An expert skill for reviewing Go repositories against Domain-Driven Design principles from Evans and Vernon.

## When to Use

- Creating a new repository interface in Go
- Reviewing existing repositories for DDD compliance
- Naming repository methods
- Deciding between Repository and Query Object
- Learning DDD repository patterns

## The 12-Point Checklist

Every repository review checks:

- [ ] One repository per aggregate root
- [ ] Interface defined in domain layer
- [ ] Implementation in infrastructure layer
- [ ] Methods reflect ubiquitous language
- [ ] Returns complete aggregates
- [ ] No business logic in repository
- [ ] Uses meaningful method names (not Find/Select/Query)
- [ ] Handles identity generation when needed
- [ ] Delegates transactions to application service
- [ ] Separate query objects for read-heavy operations
- [ ] Mockable for unit testing
- [ ] Respects aggregate boundaries

## How I Review Your Code

When you provide a repository, I will:

1. **Check all 12 principles** against your code
2. **Identify violations** with severity levels:
   - 🔴 **CRITICAL** - Breaks core DDD (must fix)
   - 🟡 **WARNING** - Breaks best practices (should fix)
   - 🟢 **SUGGESTION** - Could improve (nice to have)
3. **Provide corrected code** with explanations
4. **Suggest method names** based on Ubiquitous Language
5. **Differentiate Repository vs Query Object**

## Correct Go Repository Pattern

### Interface (Domain Layer)
```go
// domain/order/repository.go
package order

import "context"

type OrderRepository interface {
	ByID(ctx context.Context, id OrderID) (*Order, error)
	ByCustomer(ctx context.Context, customerID CustomerID) ([]*Order, error)
	AllPending(ctx context.Context) ([]*Order, error)
	Add(ctx context.Context, order *Order) error
	Remove(ctx context.Context, id OrderID) error
}
```

### Implementation (Infrastructure Layer)
```go
// infrastructure/persistence/postgres_order_repository.go
package persistence

import (
	"context"
	"database/sql"
	"myapp/domain/order"
)

type PostgresOrderRepository struct {
	db *sql.DB
}

func (r *PostgresOrderRepository) ByID(ctx context.Context, id order.OrderID) (*order.Order, error) {
	query := `SELECT id, customer_id FROM orders WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id.Value())
	// Implementation...
}

func (r *PostgresOrderRepository) Add(ctx context.Context, o *order.Order) error {
	query := `INSERT INTO orders VALUES (...)`
	return r.db.ExecContext(ctx, query, ...)
}
```

## Critical Anti-Patterns

### ❌ Multiple Aggregates in One Repository
```go
// WRONG
type SalesRepository interface {
	FindOrder(id string) (*Order, error)
	FindCustomer(id string) (*Customer, error)
	FindInvoice(id string) (*Invoice, error)
}

// CORRECT - Create separate repositories
type OrderRepository interface { }
type CustomerRepository interface { }
type InvoiceRepository interface { }
```

### ❌ Generic Method Names
```go
// WRONG
Find(id string)
Get(id string)
Save(entity *Order)
Delete(id string)

// CORRECT
ByID(ctx context.Context, id OrderID)
Add(ctx context.Context, order *Order)
Remove(ctx context.Context, id OrderID)
```

### ❌ Business Logic in Repository
```go
// WRONG
func (r *Repo) AddWithTax(order *Order, calc TaxCalculator) error {
	tax := calc.Calculate(order)
	order.SetTax(tax)  // Business logic!
	return r.persist(order)
}

// CORRECT - Logic in domain, persistence in repo
func (r *Repo) Add(ctx context.Context, order *Order) error {
	// Just persist
}
```

### ❌ Repository Manages Transactions
```go
// WRONG
func (r *Repo) AddAndNotify(order *Order, notifier Notifier) error {
	tx := r.db.BeginTx(ctx, nil)
	r.save(tx, order)
	notifier.Notify()
	tx.Commit()  // Repository manages transaction!
}

// CORRECT - Application Service manages
type Service struct {
	repo order.OrderRepository
	notifier Notifier
	db *sql.DB
}

func (s *Service) Execute(ctx context.Context, cmd Command) error {
	tx, _ := s.db.BeginTx(ctx, nil)
	s.repo.Add(ctx, order)
	s.notifier.Notify()
	tx.Commit()
}
```

### ❌ Returns Entities Instead of Aggregates
```go
// WRONG - Breaks aggregate boundary
type OrderLineItemRepository interface {
	ByOrderID(ctx context.Context, orderID string) ([]*LineItem, error)
}

// CORRECT - Load entire aggregate
type OrderRepository interface {
	ByID(ctx context.Context, id OrderID) (*Order, error)
}
```

## Method Naming Conventions

### ✅ Correct Pattern
```go
// By [Criteria]
ByID(ctx context.Context, id OrderID)
ByCustomer(ctx context.Context, customerID CustomerID)
ByEmail(ctx context.Context, email string)

// All [Status]
AllPending(ctx context.Context)
AllActive(ctx context.Context)
AllArchived(ctx context.Context)

// Specific domain queries
OverdueOrders(ctx context.Context)
WithinDateRange(ctx context.Context, start, end time.Time)
```

### ❌ Avoid These Names
```go
Find()        // Too generic
Get()         // Too generic
Select()      // Database language
Query()       // Technical, not domain
Fetch()       // Vague
Execute()     // Unclear
```

## Repository vs Query Object

### Use Repository When:
Loading aggregates to modify
```go
order, _ := orderRepo.ByID(ctx, orderId)
order.Ship()
orderRepo.Add(ctx, order)
```

### Use Query Object When:
Read-only reporting queries
```go
type OrderQueries interface {
	CustomerSalesReport(ctx context.Context, customerID CustomerID) 
		([]OrderReportDTO, error)
}
```

**Key Difference:**
- **Repository** = Load aggregates to modify (commands)
- **Query Object** = Read-only projections for reporting (queries)

## Feedback Format

### When Code Is Good ✅
```
✅ PASS: [Reason]
✅ PASS: [Reason]
🟢 SUGGESTION: [Optional improvement]
```

### When Code Has Issues 🔴🟡
```
🔴 CRITICAL: [Core principle violated]
   Issue: [Explanation]
   Fix: [Corrected code]
   
🟡 WARNING: [Best practice broken]
   Current: [Bad example]
   Better: [Good example]
```

## Ready to Review?

Just provide your repository code:

```
Here's my OrderRepository:

[paste interface and/or implementation]
```

I'll check all 12 DDD principles and give you expert feedback.

## Key References

Based on:
- **Eric Evans** - "Domain-Driven Design: Tackling Complexity in the Heart of Software"
- **Vaughn Vernon** - "Implementing Domain-Driven Design"

Core principle: Repository is an abstraction that makes the entire aggregate appear as if it were in memory.