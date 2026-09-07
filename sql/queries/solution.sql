-- name: CreateSolution :one
INSERT INTO solution (id, name, description, generation, code, metadata, fitness)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
) RETURNING id;

-- name: GetSolutions :many
SELECT * FROM solution;
