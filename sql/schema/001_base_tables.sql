-- +goose Up

CREATE TABLE run(
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    seed INT NOT NULL
);

CREATE TABLE experiment(
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NOT NULL
);

CREATE TABLE tag(
    id UUID PRIMARY KEY,
    tag TEXT NOT NULL UNIQUE
);

CREATE TABLE problem (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    hash TEXT NOT NULL,
    prompt TEXT NOT NULL,
    evaluator TEXT NOT NULL,
    minimisation BOOL NOT NULL,
    config JSONB,
    CONSTRAINT unique_problem UNIQUE (hash)
);

CREATE TABLE solution(
    id UUID PRIMARY KEY,
    name TEXT,
    description TEXT,
    generation INT,
    code TEXT,
    metadata JSONB,
    fitness JSONB
);

CREATE TABLE method(
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    hash TEXT NOT NULL,
    source TEXT NOT NULL,
    config JSONB NOT NULL,
    CONSTRAINT unique_method UNIQUE(hash)
);

CREATE TABLE message(
    id UUID PRIMARY KEY,
    message TEXT NOT NULL,
    hash BYTEA UNIQUE NOT NULL
);

CREATE TABLE llm(
    id UUID PRIMARY KEY,
    model TEXT NOT NULL,
    hash TEXT NOT NULL,
    hardware JSONB,
    config JSONB NOT NULL,
    CONSTRAINT unique_llm UNIQUE(hash)
);

-- +goose Down

DROP TABLE run;
DROP TABLE experiment;
DROP TABLE tag;
DROP TABLE problem;
DROP TABLE solution;
DROP TABLE method;
DROP TABLE message;
DROP TABLE llm;
