-- name: CreateRun :one
INSERT INTO run (id, seed)
VALUES (
    $1,
    $2
)
RETURNING id;

-- name: GetRuns :many
SELECT * FROM run;
