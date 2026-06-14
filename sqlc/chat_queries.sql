-- name: ConversationDelete :exec
DELETE FROM conversations WHERE id = $1;

-- name: ConversationGet :one
SELECT id, title, created_at, updated_at FROM conversations WHERE id = $1;

-- name: ConversationInsertOne :one
INSERT INTO conversations (title)
VALUES ($1)
RETURNING id, title, created_at, updated_at;

-- name: ConversationTouch :exec
UPDATE conversations SET updated_at = now() WHERE id = $1;

-- name: ConversationsList :many
SELECT id, title, updated_at FROM conversations ORDER BY updated_at DESC;

-- name: TurnInsertOne :exec
INSERT INTO turns (conversation_id, role, content, tool_calls, tool_call_id)
VALUES ($1, $2, $3, $4, $5);

-- name: TurnsByConversation :many
SELECT id, conversation_id, role, content, tool_calls, tool_call_id, created_at
FROM turns
WHERE conversation_id = $1
ORDER BY id;
