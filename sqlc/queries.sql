-- name: GamesInsertOne :one
INSERT INTO games (file, duration, map, mode, timestamp, type)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;

-- name: GamesDeleteOne :exec
DELETE FROM games WHERE file = $1;

-- name: TeamsInsertOne :one
INSERT INTO teams (game_id, number, result)
VALUES ($1, $2, $3)
RETURNING id;

-- name: PlayersInsertOne :one
INSERT INTO players (team_id, number, apm, color, control, mmr, name, observe, races_assigned, races_selected)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id;

-- name: MessagesInsertMany :batchexec
INSERT INTO messages (player_id, time, recipient_id, string)
VALUES ($1, $2, $3, $4);

-- name: StatsInsertMany :batchexec
INSERT INTO stats
(
    player_id,
    time,
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
VALUES
(
    $1,  -- player_id
    $2,  -- time
    $3,  -- food_made
    $4,  -- food_used
    $5,  -- minerals_collection_rate
    $6,  -- minerals_current
    $7,  -- minerals_friendly_fire_army
    $8,  -- minerals_friendly_fire_economy
    $9,  -- minerals_friendly_fire_technology
    $10, -- minerals_killed_army
    $11, -- minerals_killed_economy
    $12, -- minerals_killed_technology
    $13, -- minerals_lost_army
    $14, -- minerals_lost_economy
    $15, -- minerals_lost_technology
    $16, -- minerals_used_active_forces
    $17, -- minerals_used_current_army
    $18, -- minerals_used_current_economy
    $19, -- minerals_used_current_technology
    $20, -- minerals_used_in_progress_army
    $21, -- minerals_used_in_progress_economy
    $22, -- minerals_used_in_progress_technology
    $23, -- vespene_collection_rate
    $24, -- vespene_current
    $25, -- vespene_friendly_fire_army
    $26, -- vespene_friendly_fire_economy
    $27, -- vespene_friendly_fire_technology
    $28, -- vespene_killed_army
    $29, -- vespene_killed_economy
    $30, -- vespene_killed_technology
    $31, -- vespene_lost_army
    $32, -- vespene_lost_economy
    $33, -- vespene_lost_technology
    $34, -- vespene_used_active_forces
    $35, -- vespene_used_current_army
    $36, -- vespene_used_current_economy
    $37, -- vespene_used_current_technology
    $38, -- vespene_used_in_progress_army
    $39, -- vespene_used_in_progress_economy
    $40, -- vespene_used_in_progress_technology
    $41  -- workers_active_count
);

-- name: UnitsInsertMany :batchexec
INSERT INTO units (player_id, time, action, name, x, y)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: Statistics :many
SELECT 'Games' AS key, COUNT(*) AS value FROM games
UNION ALL
SELECT 'Teams' AS key, COUNT(*) AS value FROM teams
UNION ALL
SELECT 'Players' AS key, COUNT(*) AS value FROM players
UNION ALL
SELECT 'Messages' AS key, COUNT(*) AS value FROM messages
UNION ALL
SELECT 'Stats' AS key, COUNT(*) AS value FROM stats
UNION ALL
SELECT 'Units' AS key, COUNT(*) AS value FROM units;
