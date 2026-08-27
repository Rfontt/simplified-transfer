package postgres

import (
	"context"
	"database/sql"
	"errors"

	"event-driven-architecture/internal/domain/account"
	"event-driven-architecture/internal/domain/user"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

// AccountRepository is the PostgreSQL implementation of the domain
// account.AccountRepository port.
type AccountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

const (
	insertAccountSQL = `INSERT INTO accounts (id, owner_id, currency, balance, status, created_at)
VALUES ($1, $2, $3, $4, $5, $6)`

	selectAccountByIDSQL = `SELECT id, owner_id, currency, balance, status, created_at
FROM accounts WHERE id = $1`

	selectAllAccountsSQL = `SELECT id, owner_id, currency, balance, status, created_at
FROM accounts ORDER BY created_at`
)

func (r *AccountRepository) Add(ctx context.Context, acc *account.Account) error {
	// uuid.UUID implements driver.Valuer; the defined ID types do not, so we
	// convert explicitly before handing values to database/sql.
	_, err := r.db.ExecContext(ctx, insertAccountSQL,
		uuid.UUID(acc.ID),
		uuid.UUID(acc.OwnerId),
		acc.Balance.Currency,
		acc.Balance.Value,
		string(acc.Status),
		acc.CreatedAt,
	)
	if err != nil {
		return translateError(err, acc.OwnerId)
	}
	return nil
}

func (r *AccountRepository) ByID(ctx context.Context, id account.AccountID) (*account.Account, error) {
	row := r.db.QueryRowContext(ctx, selectAccountByIDSQL, uuid.UUID(id))
	acc, err := scanAccount(row)
	if err != nil {
		return nil, err
	}
	return acc, nil
}

func (r *AccountRepository) AllAccounts(ctx context.Context) ([]*account.Account, error) {
	rows, err := r.db.QueryContext(ctx, selectAllAccountsSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := make([]*account.Account, 0)
	for rows.Next() {
		acc, err := scanAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return accounts, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanAccount(s rowScanner) (*account.Account, error) {
	var (
		acc     account.Account
		id      uuid.UUID
		ownerID uuid.UUID
		status  string
	)
	if err := s.Scan(&id, &ownerID, &acc.Balance.Currency, &acc.Balance.Value, &status, &acc.CreatedAt); err != nil {
		return nil, err
	}
	acc.ID = account.AccountID(id)
	acc.OwnerId = user.ID(ownerID)
	acc.Status = account.AccountStatus(status)
	return &acc, nil
}

// translateError maps PostgreSQL constraint violations to domain errors so the
// calling application layer (and adapters above it) stays DB-agnostic.
func translateError(err error, ownerID user.ID) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}
	switch pgErr.Code {
	case "23505": // unique_violation
		return &account.AccountAlreadyExistsError{OwnerID: uuid.UUID(ownerID).String()}
	case "23503": // foreign_key_violation
		return &account.OwnerNotFoundError{OwnerID: uuid.UUID(ownerID).String()}
	default:
		return err
	}
}
