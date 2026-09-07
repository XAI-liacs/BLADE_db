-- name: CreateProblem :one
INSERT INTO problem (id, name, prompt, evaluator, minimisation, config)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING id;

-- name: GetProblems :many
SELECT * FROM problem;
