-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    password_hash BYTEA NOT NULL,
    default_visibility BOOL NOT NULL DEFAULT FALSE,
    is_admin BOOL NOT NULL DEFAULT FALSE
);

CREATE TABLE user_experiment(
    user_id UUID NOT NULL REFERENCES users(id),
    experiment_id UUID NOT NULL REFERENCES experiment(id),
    CONSTRAINT unique_ownership UNIQUE (user_id, experiment_id)
);

-- +goose Down
DROP TABLE user_experiment;
DROP TABLE users;
