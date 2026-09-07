-- name: CreateMethod :one
INSERT INTO method (id, name, source, config)
VALUES (
    $1,
    $2,
    $3,
    $4
) RETURNING id;

-- name: GetMethods :many
SELECT * FROM method;
