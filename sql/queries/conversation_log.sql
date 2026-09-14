-- name: CreateConversationLog :exec
INSERT INTO conversation_log (
    run_id,
    message_id,
    method_id,
    llm_id,
    created_at
)
VALUES (
    $1, $2, $3, $4, $5
);

-- name: GetConversationLog :many
SELECT
    cl.created_at,

    CASE
        WHEN cl.llm_id IS NOT NULL THEN l.model
        ELSE m.name
    END AS actor_name,

    CASE
        WHEN cl.llm_id IS NOT NULL THEN cl.llm_id
        ELSE cl.method_id
    END AS actor_id,

    CASE
        WHEN cl.llm_id IS NOT NULL THEN 'llm'
        ELSE 'method'
    END AS actor_type,

    msg.message

FROM conversation_log AS cl

JOIN message AS msg
    ON msg.id = cl.message_id

LEFT JOIN llm AS l
    ON l.id = cl.llm_id

LEFT JOIN method AS m
    ON m.id = cl.method_id

WHERE cl.run_id = $1

ORDER BY cl.created_at;
