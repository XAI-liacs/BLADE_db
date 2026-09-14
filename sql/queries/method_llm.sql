-- name: CreateMethodLLM :exec
INSERT INTO method_llm (method_id, llm_id)
VALUES (
    $1,
    $2
);

-- name: GetLLMsForMethod :many
SELECT llm.*
FROM method_llm ml
JOIN llm
    ON llm.id = ml.llm_id
WHERE ml.method_id = $1;
