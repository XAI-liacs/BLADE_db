-- name: CreateExperiment :one
INSERT INTO experiment (id, name, start_date, end_date)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING id;

-- name: GetExperiments :many
SELECT * FROM experiment;
