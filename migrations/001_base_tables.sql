-- +goose Up
CREATE TABLE Experiment(
    id UUID PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL
);

CREATE TABLE Run(
    id UUID PRIMARY KEY,
    seed INTEGER NOT NULL
);

CREATE TABLE Tag(
    id UUID PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE Message(
    id UUID PRIMARY KEY,
    message TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL
);

CREATE TABLE Problem(
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    prompt TEXT NOT NULL,
    evaluator TEXT NOT NULL,
    minimisation BOOLEAN NOT NULL,
    config JSONB
);

CREATE TABLE Method(
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    source TEXT NOT NULL,
    config JSONB NOT NULL
);

CREATE TABLE LLM(
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    hardware JSONB,
    config JSONB,
    UNIQUE (name, hardware, config)
);

CREATE TABLE Solution(
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    code TEXT NOT NULL,
    generation INTEGER,
    metadata JSONB,
    fitness JSONB NOT NULL
);

-- +goose Down
DROP TABLE Experiment;
DROP TABLE Run;
DROP TABLE Tag;
DROP TABLE Message;
DROP TABLE Problem;
DROP TABLE Method;
DROP TABLE LLM;
DROP TABLE Solution;
