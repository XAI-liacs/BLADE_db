-- name: CreateRun :one
INSERT INTO run (id, name, seed)
VALUES (
    $1,
    $2,
    $3
)
RETURNING id;

-- name: GetRuns :many
SELECT * FROM run;
