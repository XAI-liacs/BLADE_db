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

-- name: CheckExperimentWithIDExists :one
SELECT EXISTS (
    SELECT 1
    FROM experiment
    WHERE id = $1
);


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

-- name: GetUsersAndVisibilityOfExperiment :one
SELECT
    array_agg(ue.user_id)::uuid[] AS user_ids,
    e.is_public
FROM experiment e
JOIN user_experiment ue
    ON ue.experiment_id = e.id
WHERE e.id = $1
GROUP BY e.id, e.is_public;

-- name: ToggleExperimentVisibility :one
UPDATE experiment
SET is_public = NOT is_public
WHERE id = $1
RETURNING is_public;
