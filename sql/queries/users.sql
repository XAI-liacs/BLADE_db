-- name: CreateUser :one
INSERT INTO users(id, name, password_hash, is_admin)
VALUES (
    $1,
    $2,
    $3,
    $4
)  RETURNING id;

-- name: GetUserNamed :one
SELECT * FROM users
WHERE name = $1
LIMIT 1;

-- name: SetDefaultVisibility :exec
UPDATE users
SET default_visibility = $2
WHERE id = $1;

-- name: UpdatePassword :exec
UPDATE users
SET password_hash = $2
WHERE id = $1;
