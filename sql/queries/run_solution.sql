-- name: CreateRunSolution :exec
INSERT INTO run_solution (run_id, solution_id)
VALUES (
    $1,
    $2
);

-- name: GetSolutionsForRun :many
SELECT s.*,
   COALESCE(
    ARRAY(
        SELECT sp.parent_id
        FROM solution_parent_child sp
        WHERE sp.child_id = s.id
    ),
    ARRAY []::uuid[]
    )::uuid[] AS parents
FROM run_solution rs
JOIN solution s
    ON s.id = rs.solution_id
WHERE rs.run_id = $1;
