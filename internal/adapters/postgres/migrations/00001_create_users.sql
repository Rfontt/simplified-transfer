-- +goose Up
CREATE TABLE users (
    id         UUID PRIMARY KEY,
    full_name  TEXT NOT NULL,
    document   TEXT NOT NULL UNIQUE,
    email      TEXT NOT NULL UNIQUE,
    password   TEXT NOT NULL,
    type       TEXT NOT NULL
);

-- +goose Down
DROP TABLE users;
