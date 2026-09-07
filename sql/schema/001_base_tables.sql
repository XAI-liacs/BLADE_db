-- +goose Up

CREATE TABLE run(
    id UUID PRIMARY KEY,
    seed INT NOT NULL
);

CREATE TABLE experiment(
    id UUID PRIMARY KEY,
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
    prompt TEXT NOT NULL,
    evaluator TEXT NOT NULL,
    minimisation BOOL NOT NULL,
    config JSONB,
    CONSTRAINT unique_problem
        UNIQUE NULLS NOT DISTINCT (
            name,
            prompt,
            evaluator,
            minimisation,
            config
        )
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
    source TEXT NOT NULL,
    config JSONB NOT NULL
);

CREATE TABLE message(
    id UUID PRIMARY KEY,
    message TEXT UNIQUE NOT NULL
);

CREATE TABLE llm(
    id UUID PRIMARY KEY,
    model TEXT NOT NULL,
    hardware JSONB,
    config JSONB NOT NULL
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
