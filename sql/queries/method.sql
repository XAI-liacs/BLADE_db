-- name: CreateMethod :one
INSERT INTO method (id, name, hash, source, config)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT ON CONSTRAINT unique_method
DO UPDATE SET
    name = EXCLUDED.name,
    source = EXCLUDED.source,
    config = EXCLUDED.config
RETURNING id;

-- name: GetMethods :many
SELECT * FROM method;

-- name: GetMethodsInfo :many
SELECT
    m.name,
    COUNT(DISTINCT m.id) AS config_count,
    COUNT(DISTINCT rd.run_id) AS usage
FROM method m
LEFT JOIN run_descriptor rd
    ON rd.method_id = m.id
GROUP BY m.name
ORDER BY
    CASE WHEN @sort_by::text = 'Name' THEN m.name END ASC,
    CASE WHEN @sort_by::text = 'Top'
        THEN COUNT(DISTINCT rd.run_id)
    END DESC,
    m.name ASC;

-- name: GetMethodStats :many
SELECT
    m.*,
    COALESCE(r.usage_count, 0) AS method_usage_count,
    COALESCE(s.total_solution_count, 0) AS total_solution_count,
    COALESCE(s.valid_solution_count, 0) AS valid_solution_count,
    COALESCE(
        ROUND(
            s.valid_solution_count * 100.0 /
            NULLIF(s.total_solution_count, 0),
            2
        ),
        0
    ) AS success_rate,
    COALESCE(l.llm_ids, ARRAY[]::uuid[]) AS llm_ids
FROM method m
LEFT JOIN (
    SELECT
        method_id,
        COUNT(DISTINCT run_id) AS usage_count
    FROM run_descriptor
    GROUP BY method_id
) r ON r.method_id = m.id
LEFT JOIN (
    SELECT
        rd.method_id,
        COUNT(sol.id) AS total_solution_count,
        COUNT(sol.id) FILTER (
            WHERE sol.fitness IS NOT NULL
        ) AS valid_solution_count
    FROM run_descriptor rd
    JOIN run_solution rs
        ON rs.run_id = rd.run_id
    JOIN solution sol
        ON sol.id = rs.solution_id
    GROUP BY rd.method_id
) s ON s.method_id = m.id
LEFT JOIN (
    SELECT
        method_id,
        ARRAY_AGG(DISTINCT llm_id) AS llm_ids
    FROM method_llm
    GROUP BY method_id
) l ON l.method_id = m.id
WHERE m.name = $1
ORDER BY
    success_rate DESC,
    m.source,
    m.hash;
