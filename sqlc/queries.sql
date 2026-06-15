-- name: ConversationsDelete :exec
DELETE FROM conversations WHERE id = $1;

-- name: ConversationsInsert :one
INSERT INTO conversations (title)
VALUES ($1)
RETURNING id, title, created_at, updated_at;

-- name: ConversationsSelectAll :many
SELECT id, title, updated_at FROM conversations ORDER BY updated_at DESC;

-- name: ConversationsSelectByID :one
SELECT id, title, created_at, updated_at FROM conversations WHERE id = $1;

-- name: ConversationsUpdate :exec
UPDATE conversations SET updated_at = now() WHERE id = $1;

-- name: FilesDeleteAll :exec
DELETE FROM files;

-- name: FilesInsert :one
INSERT INTO files (path, parser_version, status)
VALUES ($1, $2, $3)
RETURNING id;

-- name: FilesSelectPaths :many
SELECT path FROM files;

-- name: GamesExistsByFingerprint :one
SELECT EXISTS(SELECT 1 FROM games WHERE fingerprint = $1);

-- name: GamesInsert :one
INSERT INTO games (file_id, amm, competitive, duration, fingerprint, map, mode, played_at, version)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id;

-- name: MessagesInsertMany :copyfrom
INSERT INTO messages (player_id, recipient, text, tick)
VALUES ($1, $2, $3, $4);

-- name: PlayersInsert :one
INSERT INTO players (game_id, apm, clan, color, mmr, name, number, race_assigned, race_selected, result, team)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id;

-- name: StatsInsertMany :copyfrom
INSERT INTO stats
(
    player_id,
    tick,
    food_made,
    food_used,
    minerals_collection_rate,
    minerals_current,
    minerals_friendly_fire_army,
    minerals_friendly_fire_economy,
    minerals_friendly_fire_technology,
    minerals_killed_army,
    minerals_killed_economy,
    minerals_killed_technology,
    minerals_lost_army,
    minerals_lost_economy,
    minerals_lost_technology,
    minerals_used_active_forces,
    minerals_used_current_army,
    minerals_used_current_economy,
    minerals_used_current_technology,
    minerals_used_in_progress_army,
    minerals_used_in_progress_economy,
    minerals_used_in_progress_technology,
    vespene_collection_rate,
    vespene_current,
    vespene_friendly_fire_army,
    vespene_friendly_fire_economy,
    vespene_friendly_fire_technology,
    vespene_killed_army,
    vespene_killed_economy,
    vespene_killed_technology,
    vespene_lost_army,
    vespene_lost_economy,
    vespene_lost_technology,
    vespene_used_active_forces,
    vespene_used_current_army,
    vespene_used_current_economy,
    vespene_used_current_technology,
    vespene_used_in_progress_army,
    vespene_used_in_progress_economy,
    vespene_used_in_progress_technology,
    workers_active_count
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41);

-- name: TurnsInsert :exec
INSERT INTO turns (conversation_id, role, content, tool_calls, tool_call_id)
VALUES ($1, $2, $3, $4, $5);

-- name: TurnsSelectByConversation :many
SELECT id, conversation_id, role, content, tool_calls, tool_call_id, created_at
FROM turns
WHERE conversation_id = $1
ORDER BY id;

-- name: UnitsInsertMany :copyfrom
INSERT INTO units (player_id, action, name, tick, x, y)
VALUES ($1, $2, $3, $4, $5, $6);
