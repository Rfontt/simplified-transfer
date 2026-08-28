-- +goose Up
CREATE TABLE accounts (
    id         UUID PRIMARY KEY,
    owner_id   UUID NOT NULL UNIQUE REFERENCES users(id),
    currency   TEXT NOT NULL,
    balance    DOUBLE PRECISION NOT NULL DEFAULT 0,
    status     TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE accounts;
