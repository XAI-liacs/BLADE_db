-- name: ConnectTagProblem :exec
INSERT INTO tag_problem (tag_id, problem_id)
VALUES (
    $1,
    $2
);

-- name: GetTagsforProblem :many
SELECT tag.tag
FROM tag_problem tp
JOIN tag
    ON tag.id = tp.tag_id
WHERE tp.problem_id = $1;
