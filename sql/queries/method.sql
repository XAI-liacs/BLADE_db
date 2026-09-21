-- name: CreateMethod :one
INSERT INTO method (id, name, hash, source, config)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT ON CONSTRAINT unique_method
DO UPDATE SET
    name = EXCLUDED.name,
    source = EXCLUDED.source,
    config = EXCLUDED.config
RETURNING id;

-- name: GetMethods :many
SELECT * FROM method;
