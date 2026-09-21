-- name: CreateMessage :one
INSERT INTO message (id, message, hash)
    VALUES ($1, $2, $3)
    ON CONFLICT (hash)
        DO UPDATE SET message = EXCLUDED.message
RETURNING id;

-- name: GetMessages :many
SELECT * FROM message;
