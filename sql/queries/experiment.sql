-- name: CreateExperiment :one
INSERT INTO experiment (id, start_date, end_date)
VALUES (
    $1,
    $2,
    $3
)
RETURNING id;

-- name: GetExperiments :many
SELECT * FROM experiment;
