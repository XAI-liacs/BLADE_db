-- name: CreateTag :one
INSERT INTO tag (id, tag)
    VALUES ($1, $2)
    ON CONFLICT (tag)
        DO UPDATE SET tag = EXCLUDED.tag
RETURNING id;

-- name: GetTags :many
SELECT * FROM tag;
