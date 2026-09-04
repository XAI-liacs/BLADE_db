-- name: ClearRun :exec
DELETE FROM run;

-- name: ClearExperiment :exec
DELETE FROM experiment;

-- name: ClearTag :exec
DELETE FROM tag;
