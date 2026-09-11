-- name: CreateRunSolution :exec
INSERT INTO run_solution (run_id, solution_id)
VALUES (
    $1,
    $2
);

-- name: GetSolutionsForRun :many
SELECT solution.*
FROM run_solution rs
JOIN solution
    ON solution.id = rs.solution_id
WHERE rs.run_id = $1;
