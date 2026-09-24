-- +goose Up
CREATE TABLE log(
    id UUID PRIMARY KEY,
    type TEXT NOT NULL,
    message TEXT NOT NULL,
    database_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE log;
