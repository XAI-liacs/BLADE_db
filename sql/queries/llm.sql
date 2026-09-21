-- name: CreateLLM :one
INSERT INTO llm (id, model, hash, hardware, config)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT ON CONSTRAINT unique_llm
DO UPDATE SET
    model = EXCLUDED.model,
    hardware = EXCLUDED.hardware,
    config = EXCLUDED.config
RETURNING id;

-- name: GetLLMs :many
SELECT * FROM llm;
