-- name: CreateExperimentRun :exec
INSERT INTO experiment_run (experiment_id, run_id)
VALUES (
    $1,
    $2
);

-- name: GetRunsForExperiment :many
SELECT  run.*
FROM experiment_run er
JOIN run
    ON run.id = er.run_id
WHERE er.experiment_id = $1;
