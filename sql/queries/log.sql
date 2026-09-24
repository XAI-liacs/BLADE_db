-- name: CreateLog :exec
INSERT INTO log (id, type, message, database_key, created_at)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
);

-- name: ClearLogBefore :exec
DELETE FROM log
WHERE created_at < $1;

-- name: GetLogs :many
SELECT * FROM log
ORDER BY created_at;
