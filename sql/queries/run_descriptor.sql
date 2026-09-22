-- name: CreateRunDescriptor :exec
INSERT INTO run_descriptor (run_id, method_id, problem_id)
VALUES (
    $1,
    $2,
    $3);

-- name: GetMethodAndProblemForRun :many
SELECT
    p.id           AS problem_id,
    p.name         AS problem_name,
    p.prompt       AS problem_prompt,
    p.evaluator    AS problem_evaluator,
    p.minimisation AS problem_minimisation,
    p.config       AS problem_config,

    m.id           AS method_id,
    m.name         AS method_name,
    m.source       AS method_source,
    m.config       AS method_config
FROM run_descriptor rd
JOIN problem p
    ON p.id = rd.problem_id
JOIN method m
    ON m.id = rd.method_id
WHERE rd.run_id = $1;
