-- name: CreateMessage :one
INSERT INTO message (id, message)
VALUES (
    $1,
    $2
)
RETURNING id;

-- name: GetMessages :many
SELECT * FROM message;
