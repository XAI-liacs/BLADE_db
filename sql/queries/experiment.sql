-- name: CreateExperiment :one
INSERT INTO experiment (id, name, start_date, end_date, is_public)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
) ON CONFLICT (id)
DO UPDATE SET id = EXCLUDED.id
RETURNING id;

-- name: GetExperiments :many
SELECT * FROM experiment;

-- name: GetExperimentsForUser :many
SELECT e.*
FROM experiment e
WHERE e.is_public
   OR EXISTS (
       SELECT 1
       FROM user_experiment er
       WHERE er.experiment_id = e.id
         AND er.user_id = $1
   );
