-- name: ChatGg :many
SELECT me.result,
       COUNT(*) AS games,
       COUNT(*) FILTER (WHERE EXISTS (SELECT 1 FROM messages m WHERE m.player_id = me.id AND LOWER(m.text) LIKE 'gg%')) AS said_gg
FROM (SELECT id, result FROM players WHERE players.name = $1) me
WHERE me.result IN ('Win', 'Loss')
GROUP BY me.result
ORDER BY me.result;

-- name: ChatPhrases :many
SELECT LOWER(TRIM(m.text)) AS phrase, COUNT(*) AS count
FROM (SELECT id FROM players WHERE players.name = $1) me
JOIN messages m ON m.player_id = me.id
GROUP BY phrase
ORDER BY count DESC, phrase
LIMIT 20;

-- name: DurationsBins :many
WITH me AS (
    SELECT game_id, result FROM players WHERE players.name = $1
)
SELECT LEAST(g.duration / 300, 8)::int AS bin,
       COUNT(*) AS games,
       COUNT(*) FILTER (WHERE m.result = 'Win') AS wins,
       COUNT(*) FILTER (WHERE m.result = 'Loss') AS losses
FROM me m
JOIN games g ON g.id = m.game_id
GROUP BY bin
ORDER BY bin;

-- name: DurationsBinsByRace :many
WITH me AS (
    SELECT game_id, result, race_assigned FROM players WHERE players.name = $1
)
SELECT m.race_assigned AS race,
       LEAST(g.duration / 300, 8)::int AS bin,
       COUNT(*) AS games,
       COUNT(*) FILTER (WHERE m.result = 'Win') AS wins,
       COUNT(*) FILTER (WHERE m.result = 'Loss') AS losses
FROM me m
JOIN games g ON g.id = m.game_id
GROUP BY race, bin
ORDER BY race, bin;

-- name: EconomyAt :many
SELECT s.tick,
       me.result,
       COUNT(*) AS games,
       ROUND(AVG(s.minerals_collection_rate + s.vespene_collection_rate))::int AS income,
       ROUND(AVG(s.workers_active_count))::int AS workers
FROM (SELECT id, result FROM players WHERE players.name = $1 AND players.race_assigned = $2) me
JOIN stats s ON s.player_id = me.id AND s.tick IN (4800, 9600, 14400)
WHERE me.result IN ('Win', 'Loss')
GROUP BY s.tick, me.result
ORDER BY s.tick, me.result;

-- name: FilesCountByStatus :many
SELECT status, COUNT(*) AS count FROM files GROUP BY status ORDER BY status;

-- name: FilesDeleteAll :exec
DELETE FROM files;

-- name: FilesInsertOne :one
INSERT INTO files (path, parser_version, status)
VALUES ($1, $2, $3)
RETURNING id;

-- name: FilesSelectPaths :many
SELECT path FROM files;

-- name: GamesCount :one
SELECT COUNT(*) FROM games;

-- name: GamesFingerprintExists :one
SELECT EXISTS(SELECT 1 FROM games WHERE fingerprint = $1);

-- name: GamesInsertOne :one
INSERT INTO games (file_id, amm, competitive, duration, fingerprint, map, mode, played_at, version)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id;

-- name: GamesSequence :many
SELECT g.played_at, me.race_assigned, me.mmr, me.result
FROM (SELECT game_id, race_assigned, mmr, result FROM players WHERE players.name = $1) me
JOIN games g ON g.id = me.game_id
ORDER BY g.played_at;

-- name: ImpactPair :many
WITH ours AS (
    SELECT DISTINCT game_id, team, result FROM players WHERE players.name = ANY($1::text[])
),
a AS (
    SELECT DISTINCT game_id, team FROM players WHERE players.name = $2
),
b AS (
    SELECT DISTINCT game_id, team FROM players WHERE players.name = $3
)
SELECT CASE
           WHEN a.team = o.team AND b.team = o.team THEN 'both our team'
           WHEN a.team <> o.team AND b.team <> o.team THEN 'both opposition'
           ELSE 'split (one each side)'
       END AS side,
       COUNT(*) AS games,
       COUNT(*) FILTER (WHERE o.result = 'Win') AS wins,
       COUNT(*) FILTER (WHERE o.result = 'Loss') AS losses
FROM ours o
JOIN a ON a.game_id = o.game_id
JOIN b ON b.game_id = o.game_id
GROUP BY side
ORDER BY side;

-- name: ImpactSingle :many
WITH ours AS (
    SELECT DISTINCT game_id, team, result FROM players WHERE players.name = ANY($1::text[])
),
target AS (
    SELECT DISTINCT game_id, team FROM players WHERE players.name = $2
)
SELECT CASE WHEN t.team = o.team THEN 'our team' ELSE 'opposition' END AS side,
       COUNT(*) AS games,
       COUNT(*) FILTER (WHERE o.result = 'Win') AS wins,
       COUNT(*) FILTER (WHERE o.result = 'Loss') AS losses
FROM ours o
JOIN target t ON t.game_id = o.game_id
GROUP BY side
ORDER BY side;

-- name: MapsSummary :many
SELECT g.map,
       COUNT(*) AS games,
       COUNT(*) FILTER (WHERE me.result = 'Win') AS wins,
       COUNT(*) FILTER (WHERE me.result = 'Loss') AS losses,
       COUNT(*) FILTER (WHERE g.played_at > now() - interval '180 days') AS recent_games,
       COUNT(*) FILTER (WHERE me.result = 'Win' AND g.played_at > now() - interval '180 days') AS recent_wins,
       COUNT(*) FILTER (WHERE me.result = 'Loss' AND g.played_at > now() - interval '180 days') AS recent_losses,
       MAX(g.played_at)::date AS last_played
FROM (SELECT game_id, result FROM players WHERE players.name = $1) me
JOIN games g ON g.id = me.game_id
GROUP BY g.map
HAVING COUNT(*) >= 10
ORDER BY games DESC;

-- name: MatchupPerGame :many
SELECT me.race_assigned AS my_race,
       g.mode,
       me.result,
       COALESCE(string_agg(LEFT(p.race_assigned, 1), '' ORDER BY p.race_assigned), '')::text AS opposition
FROM (SELECT game_id, team, race_assigned, result FROM players WHERE players.name = $1) me
JOIN games g ON g.id = me.game_id
JOIN players p ON p.game_id = me.game_id AND p.team <> me.team
GROUP BY me.game_id, me.race_assigned, g.mode, me.result;

-- name: MaxoutBuckets :many
WITH me AS (
    SELECT id, result FROM players WHERE players.name = $1 AND players.race_assigned = $2
),
supply AS (
    SELECT me.id, me.result, MAX(s.food_used) AS peak
    FROM me
    JOIN stats s ON s.player_id = me.id
    GROUP BY me.id, me.result
),
built AS (
    SELECT me.id, COUNT(*) AS built
    FROM me
    JOIN units u ON u.player_id = me.id AND u.action IN ('born', 'init') AND u.name = $3
    GROUP BY me.id
)
SELECT CASE
           WHEN s.peak < 195 THEN 'not maxed'
           WHEN b.built IS NULL THEN 'maxed, none'
           WHEN b.built < 4 THEN 'maxed, 1-3'
           WHEN b.built < 8 THEN 'maxed, 4-7'
           ELSE 'maxed, 8+'
       END AS scenario,
       COUNT(*) AS games,
       COUNT(*) FILTER (WHERE s.result = 'Win') AS wins,
       COUNT(*) FILTER (WHERE s.result = 'Loss') AS losses
FROM supply s
LEFT JOIN built b ON b.id = s.id
GROUP BY scenario
ORDER BY scenario;

-- name: MaxoutLossComposition :many
WITH me AS (
    SELECT id, game_id, team, result FROM players WHERE players.name = $1 AND players.race_assigned = $2
),
loss_games AS (
    SELECT me.game_id, me.team
    FROM me
    JOIN stats s ON s.player_id = me.id
    WHERE me.result = 'Loss'
      AND EXISTS (SELECT 1 FROM units u WHERE u.player_id = me.id AND u.action IN ('born', 'init') AND u.name = $3)
    GROUP BY me.game_id, me.team
    HAVING MAX(s.food_used) >= 195
)
SELECT u.name, COUNT(*) AS built
FROM loss_games lg
JOIN players p ON p.game_id = lg.game_id AND p.team <> lg.team
JOIN units u ON u.player_id = p.id AND u.action IN ('born', 'init')
WHERE u.name NOT IN ('AutoTurret', 'Broodling', 'CreepTumor', 'CreepTumorQueen', 'Drone', 'Egg', 'Interceptor', 'Larva', 'MULE', 'Overlord', 'Overseer', 'Probe', 'SCV')
GROUP BY u.name
ORDER BY built DESC
LIMIT 15;

-- name: MaxoutLossMaps :many
WITH me AS (
    SELECT id, game_id, team, result FROM players WHERE players.name = $1 AND players.race_assigned = $2
),
loss_games AS (
    SELECT me.game_id, me.team
    FROM me
    JOIN stats s ON s.player_id = me.id
    WHERE me.result = 'Loss'
      AND EXISTS (SELECT 1 FROM units u WHERE u.player_id = me.id AND u.action IN ('born', 'init') AND u.name = $3)
    GROUP BY me.game_id, me.team
    HAVING MAX(s.food_used) >= 195
)
SELECT g.map, COUNT(*) AS losses
FROM loss_games lg
JOIN games g ON g.id = lg.game_id
GROUP BY g.map
ORDER BY losses DESC
LIMIT 10;

-- name: MaxoutLossOpponents :many
WITH me AS (
    SELECT id, game_id, team, result FROM players WHERE players.name = $1 AND players.race_assigned = $2
),
loss_games AS (
    SELECT me.game_id, me.team
    FROM me
    JOIN stats s ON s.player_id = me.id
    WHERE me.result = 'Loss'
      AND EXISTS (SELECT 1 FROM units u WHERE u.player_id = me.id AND u.action IN ('born', 'init') AND u.name = $3)
    GROUP BY me.game_id, me.team
    HAVING MAX(s.food_used) >= 195
)
SELECT p.name, p.race_assigned AS race, COUNT(*) AS games, ROUND(AVG(p.mmr))::int AS avg_mmr
FROM loss_games lg
JOIN players p ON p.game_id = lg.game_id AND p.team <> lg.team
GROUP BY p.name, p.race_assigned
HAVING COUNT(*) >= 2
ORDER BY games DESC
LIMIT 10;

-- name: MessagesCount :one
SELECT COUNT(*) FROM messages;

-- name: MessagesInsertMany :copyfrom
INSERT INTO messages (player_id, recipient, text, tick)
VALUES ($1, $2, $3, $4);

-- name: OpeningsPerGame :many
SELECT me.game_id,
       me.result,
       g.map,
       EXTRACT(YEAR FROM g.played_at)::int AS year,
       COALESCE(MIN(CASE WHEN u.name = $3 THEN u.tick END), 2147483647)::int AS first_a,
       COALESCE(MIN(CASE WHEN u.name = $4 THEN u.tick END), 2147483647)::int AS first_b
FROM (SELECT id, game_id, result FROM players WHERE players.name = $1 AND players.race_assigned = $2) me
JOIN units u ON u.player_id = me.id AND u.action IN ('born', 'init')
JOIN games g ON g.id = me.game_id
GROUP BY me.game_id, me.result, g.map, year;

-- name: PartnersPerGame :many
SELECT COALESCE(string_agg(p.name, ' + ' ORDER BY p.name), '')::text AS lineup,
       COALESCE(MIN(p.result), '')::text AS result
FROM players p
WHERE p.name = ANY($1::text[])
GROUP BY p.game_id;

-- name: PlayersCount :one
SELECT COUNT(*) FROM players;

-- name: PlayersInsertOne :one
INSERT INTO players (game_id, apm, clan, color, mmr, name, number, race_assigned, race_selected, result, team)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id;

-- name: Report1 :many
SELECT played_at, map, mode, duration, winners, losers
FROM matches
ORDER BY played_at;

-- name: Report2 :many
SELECT date(g.played_at) AS game_date,
       COUNT(DISTINCT g.id) FILTER (WHERE p.result = 'Win') AS wins,
       COUNT(DISTINCT g.id) FILTER (WHERE p.result = 'Loss') AS losses
FROM games g
JOIN players p ON p.game_id = g.id
WHERE p.name = ANY($1::text[])
GROUP BY date(g.played_at)
ORDER BY date(g.played_at);

-- name: RivalsTop :many
SELECT p.name,
       COUNT(*) AS games,
       COUNT(*) FILTER (WHERE me.result = 'Win') AS wins,
       COUNT(*) FILTER (WHERE me.result = 'Loss') AS losses,
       ROUND(AVG(me.mmr))::int AS my_mmr,
       ROUND(AVG(p.mmr))::int AS their_mmr
FROM (SELECT game_id, team, mmr, result FROM players WHERE players.name = $1) me
JOIN players p ON p.game_id = me.game_id AND p.team <> me.team
GROUP BY p.name
HAVING COUNT(*) >= 30
ORDER BY games DESC
LIMIT 20;

-- name: StatsCount :one
SELECT COUNT(*) FROM stats;

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

-- name: UnitsCount :one
SELECT COUNT(*) FROM units;

-- name: UnitsInsertMany :copyfrom
INSERT INTO units (player_id, action, name, tick, x, y)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: VersusByMap :many
WITH me AS (
    SELECT game_id, team, result FROM players WHERE players.name = $1
),
other AS (
    SELECT game_id, team FROM players WHERE players.name = $2
)
SELECT g.map,
       COUNT(*) AS games,
       COUNT(*) FILTER (WHERE m.result = 'Win') AS wins,
       COUNT(*) FILTER (WHERE m.result = 'Loss') AS losses
FROM me m
JOIN other o ON o.game_id = m.game_id AND o.team <> m.team
JOIN games g ON g.id = m.game_id
GROUP BY g.map
HAVING COUNT(*) >= 4
ORDER BY games DESC;

-- name: VersusByRace :many
WITH me AS (
    SELECT game_id, team, result, mmr, apm FROM players WHERE players.name = $1
),
other AS (
    SELECT game_id, team, mmr, apm, race_assigned FROM players WHERE players.name = $2
)
SELECT o.race_assigned AS race,
       g.mode,
       COUNT(*) AS games,
       COUNT(*) FILTER (WHERE m.result = 'Win') AS wins,
       COUNT(*) FILTER (WHERE m.result = 'Loss') AS losses,
       ROUND(AVG(m.mmr))::int AS my_mmr,
       ROUND(AVG(o.mmr))::int AS their_mmr,
       ROUND(AVG(m.apm))::int AS my_apm,
       ROUND(AVG(o.apm))::int AS their_apm
FROM me m
JOIN other o ON o.game_id = m.game_id AND o.team <> m.team
JOIN games g ON g.id = m.game_id
GROUP BY race, g.mode
ORDER BY games DESC;

-- name: VersusByYear :many
WITH me AS (
    SELECT game_id, team, result FROM players WHERE players.name = $1
),
other AS (
    SELECT game_id, team FROM players WHERE players.name = $2
)
SELECT EXTRACT(YEAR FROM g.played_at)::int AS year,
       COUNT(*) AS games,
       COUNT(*) FILTER (WHERE m.result = 'Win') AS wins,
       COUNT(*) FILTER (WHERE m.result = 'Loss') AS losses
FROM me m
JOIN other o ON o.game_id = m.game_id AND o.team <> m.team
JOIN games g ON g.id = m.game_id
GROUP BY year
ORDER BY year;

-- name: VersusRecent :many
WITH me AS (
    SELECT game_id, team, result FROM players WHERE players.name = $1
),
other AS (
    SELECT game_id, team FROM players WHERE players.name = $2
)
SELECT g.played_at, g.map, g.mode, m.result
FROM me m
JOIN other o ON o.game_id = m.game_id AND o.team <> m.team
JOIN games g ON g.id = m.game_id
ORDER BY g.played_at DESC
LIMIT 10;

-- name: VersusRelation :many
WITH me AS (
    SELECT game_id, team, result FROM players WHERE players.name = $1
),
other AS (
    SELECT game_id, team FROM players WHERE players.name = $2
)
SELECT CASE WHEN o.team = m.team THEN 'teammate' ELSE 'opposing' END AS relation,
       COUNT(*) AS games,
       COUNT(*) FILTER (WHERE m.result = 'Win') AS wins,
       COUNT(*) FILTER (WHERE m.result = 'Loss') AS losses
FROM me m
JOIN other o ON o.game_id = m.game_id
GROUP BY relation
ORDER BY relation DESC;
