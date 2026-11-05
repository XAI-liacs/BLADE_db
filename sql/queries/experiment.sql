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

-- name: GetAllExperimentsForUser :many
SELECT e.*
FROM experiment e
WHERE e.is_public
   OR EXISTS (
       SELECT 1
       FROM user_experiment ue
       WHERE ue.experiment_id = e.id
         AND ue.user_id = $1
   )
   ORDER BY
       CASE WHEN sqlc.arg(sort_column)::text = 'start_date'
           THEN e.start_date END,
       CASE WHEN sqlc.arg(sort_column)::text = 'name'
           THEN e.name END;

-- name: GetPublicExperiments :many
SELECT *
FROM experiment e
WHERE e.is_public
ORDER BY
    CASE WHEN sqlc.arg(sort_column)::text = 'start_date'
        THEN e.start_date END,
    CASE WHEN sqlc.arg(sort_column)::text = 'name'
        THEN e.name END;

-- name: GetPrivateExperimentsForUser :many
SELECT e.* FROM user_experiment ue
JOIN experiment e
ON e.id = ue.experiment_id
WHERE ue.user_id = $1
ORDER BY
    CASE WHEN sqlc.arg(sort_column)::text = 'start_date'
        THEN e.start_date END,
    CASE WHEN sqlc.arg(sort_column)::text = 'name'
        THEN e.name END;

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
