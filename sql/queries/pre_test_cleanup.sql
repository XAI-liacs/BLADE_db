-- name: ClearRun :exec
DELETE FROM run;

-- name: ClearExperiment :exec
DELETE FROM experiment;

-- name: ClearTag :exec
DELETE FROM tag;

-- name: ClearMessage :exec
DELETE FROM message;

-- name: ClearProblem :exec
DELETE FROM problem;

-- name: ClearLLM :exec
DELETE FROM llm;

-- name: ClearMethod :exec
DELETE FROM method;

-- name: ClearSolution :exec
DELETE FROM solution;

-- name: ClearTagProblem :exec
DELETE FROM tag_problem;

-- name: ClearRunSolution :exec
DELETE FROM run_solution;
