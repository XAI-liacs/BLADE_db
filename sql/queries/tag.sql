-- name: CreateTag :one
INSERT INTO tag (id, tag)
VALUES (
    $1,
    $2
)
RETURNING id;

-- name: GetTags :many
SELECT * FROM tag;
