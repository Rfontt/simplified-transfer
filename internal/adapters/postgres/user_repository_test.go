package postgres

import (
	"context"
	"errors"
	"testing"

	"event-driven-architecture/internal/domain/user"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func testUser(t *testing.T) (*user.User, uuid.UUID) {
	t.Helper()
	id := uuid.New()
	u := &user.User{
		ID:           user.ID(id),
		FullName:     user.FullName("Rita Fontenele"),
		Document:     user.Document("52998224725"),
		Email:        user.Email("rita@example.com"),
		PasswordHash: "$2a$10$hashed",
		Type:         user.COMMON,
	}
	return u, id
}

func TestUserRepository_Add(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewUserRepository(db)
	u, id := testUser(t)

	mock.ExpectExec(`INSERT INTO users`).
		WithArgs(id.String(), "Rita Fontenele", "52998224725", "rita@example.com", "$2a$10$hashed", "common").
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.Add(context.Background(), u); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepository_AddDocumentConflict(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewUserRepository(db)
	u, _ := testUser(t)

	mock.ExpectExec(`INSERT INTO users`).
		WillReturnError(&pgconn.PgError{Code: "23505", ConstraintName: "users_document_key"})

	err := repo.Add(context.Background(), u)
	var alreadyExists *user.AlreadyExistsError
	if !errors.As(err, &alreadyExists) {
		t.Fatalf("expected AlreadyExistsError, got %v", err)
	}
	if alreadyExists.Document == "" || alreadyExists.Email != "" {
		t.Errorf("expected document conflict, got %+v", alreadyExists)
	}
}

func TestUserRepository_AddEmailConflict(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewUserRepository(db)
	u, _ := testUser(t)

	mock.ExpectExec(`INSERT INTO users`).
		WillReturnError(&pgconn.PgError{Code: "23505", ConstraintName: "users_email_key"})

	err := repo.Add(context.Background(), u)
	var alreadyExists *user.AlreadyExistsError
	if !errors.As(err, &alreadyExists) {
		t.Fatalf("expected AlreadyExistsError, got %v", err)
	}
	if alreadyExists.Email == "" || alreadyExists.Document != "" {
		t.Errorf("expected email conflict, got %+v", alreadyExists)
	}
}

func TestUserRepository_ByID(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewUserRepository(db)
	_, id := testUser(t)

	rows := sqlmock.NewRows([]string{"id", "full_name", "document", "email", "password", "type"}).
		AddRow(id.String(), "Rita Fontenele", "52998224725", "rita@example.com", "$2a$10$hashed", "common")

	mock.ExpectQuery(`SELECT id, full_name, document, email, password, type`).
		WithArgs(id.String()).
		WillReturnRows(rows)

	u, err := repo.ByID(context.Background(), user.ID(id))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if uuid.UUID(u.ID) != id {
		t.Errorf("expected id %v, got %v", id, u.ID)
	}
	if u.FullName != "Rita Fontenele" || u.Email != "rita@example.com" {
		t.Errorf("unexpected user: %+v", u)
	}
	if u.Type != user.COMMON || u.PasswordHash != "$2a$10$hashed" {
		t.Errorf("unexpected user: %+v", u)
	}
}
