-- name: CreateMessage :one
INSERT INTO message (id, message)
    VALUES ($1, $2)
    ON CONFLICT (message)
        DO UPDATE SET message = EXCLUDED.message
RETURNING id;

-- name: GetMessages :many
SELECT * FROM message;
