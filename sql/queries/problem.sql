-- name: CreateProblem :one
INSERT INTO problem (id, name, prompt, evaluator, minimisation, config)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT ON CONSTRAINT unique_problem
DO UPDATE SET
    name = EXCLUDED.name,
    prompt = EXCLUDED.prompt,
    evaluator = EXCLUDED.evaluator,
    minimisation = EXCLUDED.minimisation,
    config = EXCLUDED.config
RETURNING id;

-- name: GetProblems :many
SELECT * FROM problem;
