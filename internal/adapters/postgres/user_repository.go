package postgres

import (
	"context"
	"database/sql"
	"errors"

	"event-driven-architecture/internal/domain/user"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

const (
	insertUserSQL = `INSERT INTO users (id, full_name, document, email, password, type)
VALUES ($1, $2, $3, $4, $5, $6)`

	selectUserByIDSQL = `SELECT id, full_name, document, email, password, type
FROM users WHERE id = $1`

	selectAllUsersSQL = `SELECT id, full_name, document, email, password, type
FROM users ORDER BY full_name`
)

func (r *UserRepository) Add(ctx context.Context, u *user.User) error {
	_, err := r.db.ExecContext(ctx, insertUserSQL,
		uuid.UUID(u.ID),
		u.FullName,
		u.Document,
		u.Email,
		u.Password,
		string(u.Type),
	)
	if err != nil {
		return translateUserError(err, u)
	}
	return nil
}

func (r *UserRepository) ByID(ctx context.Context, id user.ID) (*user.User, error) {
	row := r.db.QueryRowContext(ctx, selectUserByIDSQL, uuid.UUID(id))
	u, err := scanUser(row)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) AllUsers(ctx context.Context) ([]*user.User, error) {
	rows, err := r.db.QueryContext(ctx, selectAllUsersSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*user.User, 0)
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func scanUser(s rowScanner) (*user.User, error) {
	var (
		u   user.User
		id  uuid.UUID
		typ string
	)
	if err := s.Scan(&id, &u.FullName, &u.Document, &u.Email, &u.Password, &typ); err != nil {
		return nil, err
	}
	u.ID = user.ID(id)
	u.Type = user.Type(typ)
	return &u, nil
}

func translateUserError(err error, u *user.User) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}
	if pgErr.Code != "23505" {
		return err
	}
	switch pgErr.ConstraintName {
	case "users_document_key":
		return &user.AlreadyExistsError{Document: u.Document}
	case "users_email_key":
		return &user.AlreadyExistsError{Email: u.Email}
	default:
		return &user.AlreadyExistsError{Email: u.Email, Document: u.Document}
	}
}
