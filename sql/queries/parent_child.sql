-- name: CreateParentChild :exec
INSERT INTO solution_parent_child (parent_id, child_id)
VALUES (
    $1,
    $2
);

-- name: GetChildrenofSolution :many
SELECT solution.*
FROM solution_parent_child pc
JOIN solution
    ON solution.id = pc.child_id
WHERE pc.parent_id = $1;

-- name: GetParentsofSolution :many
SELECT solution.*
FROM solution_parent_child pc
JOIN solution
    ON solution.id = pc.parent_id
WHERE pc.child_id = $1;
