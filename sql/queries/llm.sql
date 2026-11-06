-- name: CreateLLM :one
INSERT INTO llm (id, model, hash, hardware, config)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT ON CONSTRAINT unique_llm
DO UPDATE SET
    model = EXCLUDED.model,
    hardware = EXCLUDED.hardware,
    config = EXCLUDED.config
RETURNING id;

-- name: GetLLMs :many
SELECT * FROM llm;

-- name: GetAllLLMs :many
SELECT
    model::text AS model,
    COUNT(hash)::bigint AS usage_count,
    (CASE
        WHEN hardware IS NULL THEN 'cloud'
        ELSE 'local'
    END)::text AS type
FROM llm
GROUP BY model, type
ORDER BY
    CASE WHEN sqlc.arg(sort)::text = 'Top' THEN COUNT(hash) END DESC,
    CASE WHEN sqlc.arg(sort)::text = 'Name' THEN model END ASC;

-- name: GetLocalLLMs :many
SELECT
    model::text AS model,
    COUNT(hash)::bigint AS usage_count,
    'local'::text AS type
FROM llm
WHERE hardware IS NOT NULL
GROUP BY model
ORDER BY
    CASE WHEN sqlc.arg(sort)::text = 'Top' THEN COUNT(hash) END DESC,
    CASE WHEN sqlc.arg(sort)::text = 'Name' THEN model END ASC;

-- name: GetCloudLLMs :many
SELECT
    model::text AS model,
    COUNT(hash)::bigint AS usage_count,
    'cloud'::text AS type
FROM llm
WHERE hardware IS NULL
GROUP BY model
ORDER BY
    CASE WHEN sqlc.arg(sort)::text = 'Top' THEN COUNT(hash) END DESC,
    CASE WHEN sqlc.arg(sort)::text = 'Name' THEN model END ASC;

-- name: GetLLMSolutionCounts :many
WITH single_llm_runs AS (
    SELECT
        cl.run_id,
        cl.llm_id
    FROM conversation_log cl
    WHERE cl.llm_id IS NOT NULL
    GROUP BY cl.run_id, cl.llm_id
    HAVING COUNT(DISTINCT cl.llm_id) = 1
),

solution_counts AS (
    SELECT
        slr.llm_id,
        COUNT(s.id) AS total_solution_count,
        COUNT(s.id) FILTER (
            WHERE s.fitness IS NOT NULL
        ) AS valid_solution_count
    FROM single_llm_runs slr
    JOIN run_solution rs
        ON rs.run_id = slr.run_id
    JOIN solution s
        ON s.id = rs.solution_id
    GROUP BY slr.llm_id
)

SELECT
    llm.*,
    COALESCE(sc.total_solution_count, 0)::BIGINT AS total_solution_count,
    COALESCE(sc.valid_solution_count, 0)::BIGINT AS valid_solution_count
FROM llm
LEFT JOIN solution_counts sc
    ON sc.llm_id = llm.id
WHERE llm.model = sqlc.arg(model);
