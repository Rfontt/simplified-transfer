package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"event-driven-architecture/internal/domain"
	"event-driven-architecture/internal/domain/account"
	"event-driven-architecture/internal/domain/user"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, mock
}

func testAccount(t *testing.T) (*account.Account, uuid.UUID, uuid.UUID) {
	t.Helper()
	id := uuid.New()
	ownerID := uuid.New()
	acc := &account.Account{
		ID:        account.AccountID(id),
		OwnerId:   user.ID(ownerID),
		Balance:   domain.MonetaryAmount{Currency: "BRL", Value: 100},
		Status:    account.ACTIVE,
		CreatedAt: time.Now(),
	}
	return acc, id, ownerID
}

func TestAccountRepository_Add(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewAccountRepository(db)
	acc, id, ownerID := testAccount(t)

	mock.ExpectExec(`INSERT INTO accounts`).
		WithArgs(id.String(), ownerID.String(), "BRL", 100.0, "active", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.Add(context.Background(), acc); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestAccountRepository_AddUniqueViolation(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewAccountRepository(db)
	acc, _, _ := testAccount(t)

	mock.ExpectExec(`INSERT INTO accounts`).
		WillReturnError(&pgconn.PgError{Code: "23505"})

	err := repo.Add(context.Background(), acc)
	var alreadyExists *account.AccountAlreadyExistsError
	if !errors.As(err, &alreadyExists) {
		t.Fatalf("expected AccountAlreadyExistsError, got %v", err)
	}
}

func TestAccountRepository_AddForeignKeyViolation(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewAccountRepository(db)
	acc, _, _ := testAccount(t)

	mock.ExpectExec(`INSERT INTO accounts`).
		WillReturnError(&pgconn.PgError{Code: "23503"})

	err := repo.Add(context.Background(), acc)
	var ownerNotFound *account.OwnerNotFoundError
	if !errors.As(err, &ownerNotFound) {
		t.Fatalf("expected OwnerNotFoundError, got %v", err)
	}
}

func TestAccountRepository_ByID(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewAccountRepository(db)
	_, id, ownerID := testAccount(t)
	createdAt := time.Now()

	rows := sqlmock.NewRows([]string{"id", "owner_id", "currency", "balance", "status", "created_at"}).
		AddRow(id.String(), ownerID.String(), "BRL", 100.0, "active", createdAt)

	mock.ExpectQuery(`SELECT id, owner_id, currency, balance, status, created_at`).
		WithArgs(id.String()).
		WillReturnRows(rows)

	acc, err := repo.ByID(context.Background(), account.AccountID(id))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if uuid.UUID(acc.ID) != id {
		t.Errorf("expected id %v, got %v", id, acc.ID)
	}
	if uuid.UUID(acc.OwnerId) != ownerID {
		t.Errorf("expected owner %v, got %v", ownerID, acc.OwnerId)
	}
	if acc.Balance.Currency != "BRL" || acc.Balance.Value != 100.0 {
		t.Errorf("unexpected balance: %+v", acc.Balance)
	}
	if acc.Status != account.ACTIVE {
		t.Errorf("expected ACTIVE, got %v", acc.Status)
	}
}

func TestAccountRepository_AllAccounts(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewAccountRepository(db)
	_, id1, ownerID1 := testAccount(t)
	_, id2, ownerID2 := testAccount(t)

	rows := sqlmock.NewRows([]string{"id", "owner_id", "currency", "balance", "status", "created_at"}).
		AddRow(id1.String(), ownerID1.String(), "BRL", 100.0, "active", time.Now()).
		AddRow(id2.String(), ownerID2.String(), "USD", 200.0, "active", time.Now())

	mock.ExpectQuery(`SELECT id, owner_id, currency, balance, status, created_at`).
		WillReturnRows(rows)

	accounts, err := repo.AllAccounts(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(accounts) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(accounts))
	}
}
