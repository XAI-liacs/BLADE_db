-- name: CreateLLM :one
INSERT INTO llm (id, model, hardware, config)
VALUES (
    $1,
    $2,
    $3,
    $4
) RETURNING id;

-- name: GetLLMs :many
SELECT * FROM llm;
