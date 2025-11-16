-- select * from games where id = 218;

-- select * from teams where game_id = 218 order by number asc;

-- select * from teams_players where team_id in (-- select id from teams where game_id = 218) order by team_id asc, number asc;

-- select * from players where id in (-- select player_id from teams_players where team_id in (select id from teams where game_id = 218)) order by name asc;

-- select * from messages where team_player_id in (select id from teams_players where team_id in (select id from teams where game_id = 218)) order by time asc;

-- select * from stats where team_player_id in (select id from teams_players where team_id in (select id from teams where game_id = 218)) order by time asc;

-- select * from units where team_player_id in (select id from teams_players where team_id in (select id from teams where game_id = 218)) order by time asc;

WITH target_games AS (
    SELECT DISTINCT g.id, g.timestamp
    FROM games g
    JOIN teams t ON t.game_id = g.id
    JOIN teams_players tp ON tp.team_id = t.id
    JOIN players p ON p.id = tp.player_id
    WHERE g.file LIKE '%/mahendra/Windows/%' AND p.name IN ('Pineapple', 'MuNi', 'SINDIOS', 'Sajoma')
    GROUP BY g.id, g.timestamp
    HAVING COUNT(DISTINCT p.name) = 4
),
team_types AS (
    SELECT
        g.id,
        t.id as team_id,
        t.result,
        CASE WHEN COUNT(CASE WHEN p.name IN ('Pineapple', 'MuNi', 'SINDIOS') THEN 1 END) = 3 THEN 1
                 WHEN COUNT(CASE WHEN p.name = 'Sajoma' THEN 1 END) = 1 THEN 2
        END as team_type
    FROM target_games tg
    JOIN games g ON g.id = tg.id
    JOIN teams t ON t.game_id = g.id
    JOIN teams_players tp ON tp.team_id = t.id
    JOIN players p ON p.id = tp.player_id
    GROUP BY g.id, t.id, t.result
    HAVING COUNT(CASE WHEN p.name IN ('Pineapple', 'MuNi', 'SINDIOS') THEN 1 END) = 3
         OR COUNT(CASE WHEN p.name = 'Sajoma' THEN 1 END) = 1
),
team_compositions AS (
    SELECT
        tt.id,
        tt.team_id,
        tt.result,
        tt.team_type,
        STRING_AGG(
            p.name || ' (' || LEFT(tp.races_assigned, 1) || ')',
            ', '
            ORDER BY
                CASE
                    WHEN tt.team_type = 1 AND p.name IN ('Pineapple', 'MuNi', 'SINDIOS') THEN 0
                    WHEN tt.team_type = 2 AND p.name = 'Sajoma' THEN 0
                    ELSE 1
                END,
                p.name
        ) as team_info
    FROM team_types tt
    JOIN teams_players tp ON tp.team_id = tt.team_id
    JOIN players p ON p.id = tp.player_id
    GROUP BY tt.id, tt.team_id, tt.result, tt.team_type
)
SELECT
    g.timestamp,
    t1.team_info as team_1,
    t1.result as team_1_result,
    t2.team_info as team_2,
    t2.result as team_2_result,
    g.map,
    TO_CHAR((g.duration || ' seconds')::interval, 'MI:SS') as duration
FROM target_games tg
JOIN games g ON g.id = tg.id
JOIN team_compositions t1 ON t1.id = g.id AND t1.team_type = 1
JOIN team_compositions t2 ON t2.id = g.id AND t2.team_type = 2
ORDER BY g.timestamp DESC;

WITH target_games AS (
    SELECT DISTINCT g.id, g.timestamp
    FROM games g
    JOIN teams t ON t.game_id = g.id
    JOIN teams_players tp ON tp.team_id = t.id
    JOIN players p ON p.id = tp.player_id
    WHERE g.file LIKE '%/mahendra/Windows/%' AND p.name IN ('Pineapple', 'MuNi', 'SINDIOS')
    GROUP BY g.id, g.timestamp
    HAVING COUNT(DISTINCT p.name) = 3
),
team_types AS (
    SELECT
        g.id,
        t.id as team_id,
        t.result,
        CASE WHEN COUNT(CASE WHEN p.name IN ('Pineapple', 'MuNi', 'SINDIOS') THEN 1 END) = 3 THEN 1
             ELSE 2
        END as team_type
    FROM target_games tg
    JOIN games g ON g.id = tg.id
    JOIN teams t ON t.game_id = g.id
    JOIN teams_players tp ON tp.team_id = t.id
    JOIN players p ON p.id = tp.player_id
    GROUP BY g.id, t.id, t.result
    HAVING COUNT(CASE WHEN p.name IN ('Pineapple', 'MuNi', 'SINDIOS') THEN 1 END) = 3
        OR COUNT(CASE WHEN p.name NOT IN ('Pineapple', 'MuNi', 'SINDIOS') THEN 1 END) > 0
),
team_compositions AS (
    SELECT
        tt.id,
        tt.team_id,
        tt.result,
        tt.team_type,
        STRING_AGG(
            p.name || ' (' || LEFT(tp.races_assigned, 1) || ')',
            ', '
            ORDER BY
                CASE
                    WHEN tt.team_type = 1 AND p.name IN ('Pineapple', 'MuNi', 'SINDIOS') THEN 0
                    ELSE 1
                END,
                p.name
        ) as team_info
    FROM team_types tt
    JOIN teams_players tp ON tp.team_id = tt.team_id
    JOIN players p ON p.id = tp.player_id
    GROUP BY tt.id, tt.team_id, tt.result, tt.team_type
)
SELECT
    g.timestamp,
    t1.team_info as team_1,
    t1.result as team_1_result,
    t2.team_info as team_2,
    t2.result as team_2_result,
    g.map,
    TO_CHAR((g.duration || ' seconds')::interval, 'MI:SS') as duration
FROM target_games tg
JOIN games g ON g.id = tg.id
JOIN team_compositions t1 ON t1.id = g.id AND t1.team_type = 1
JOIN team_compositions t2 ON t2.id = g.id AND t2.team_type = 2
ORDER BY g.timestamp DESC;

WITH target_games AS (
    SELECT DISTINCT g.id, g.timestamp
    FROM games g
    JOIN teams t ON t.game_id = g.id
    JOIN teams_players tp ON tp.team_id = t.id
    JOIN players p ON p.id = tp.player_id
    WHERE g.file LIKE '%/mahendra/Windows/%' AND p.name IN ('Pineapple', 'MuNi', 'SINDIOS')
    GROUP BY g.id, g.timestamp
    HAVING COUNT(DISTINCT p.name) = 3
),
team_types AS (
    SELECT
        g.id,
        t.id as team_id,
        t.result,
        CASE WHEN COUNT(CASE WHEN p.name IN ('Pineapple', 'MuNi', 'SINDIOS') THEN 1 END) = 3 THEN 1
             ELSE 2
        END as team_type
    FROM target_games tg
    JOIN games g ON g.id = tg.id
    JOIN teams t ON t.game_id = g.id
    JOIN teams_players tp ON tp.team_id = t.id
    JOIN players p ON p.id = tp.player_id
    GROUP BY g.id, t.id, t.result
    HAVING COUNT(CASE WHEN p.name IN ('Pineapple', 'MuNi', 'SINDIOS') THEN 1 END) = 3
        OR COUNT(CASE WHEN p.name NOT IN ('Pineapple', 'MuNi', 'SINDIOS') THEN 1 END) > 0
),
team_compositions AS (
    SELECT
        tt.id as game_id,
        tt.team_id,
        tt.result,
        tt.team_type,
        STRING_AGG(
            p.name || ' (' || LEFT(tp.races_assigned, 1) || ')',
            ', '
            ORDER BY
                CASE
                    WHEN tt.team_type = 1 AND p.name IN ('Pineapple', 'MuNi', 'SINDIOS') THEN 0
                    ELSE 1
                END,
                p.name
        ) as team_info
    FROM team_types tt
    JOIN teams_players tp ON tp.team_id = tt.team_id
    JOIN players p ON p.id = tp.player_id
    GROUP BY tt.id, tt.team_id, tt.result, tt.team_type
),
game_results AS (
    SELECT
        DATE_TRUNC('day', g.timestamp) as date,
        CASE
            WHEN t1.result = 'Win' THEN 1
            WHEN t1.result = 'Loss' THEN 0
            ELSE NULL
        END as result
    FROM target_games tg
    JOIN games g ON g.id = tg.id
    JOIN team_compositions t1 ON t1.game_id = g.id AND t1.team_type = 1
),
cumulative_stats AS (
    SELECT
        date,
        result,
        SUM(CASE WHEN result = 1 THEN 1 ELSE 0 END) OVER (ORDER BY date) as cum_wins,
        SUM(CASE WHEN result = 0 THEN 1 ELSE 0 END) OVER (ORDER BY date) as cum_losses,
        ROUND(
            100.0 * SUM(CASE WHEN result = 1 THEN 1 ELSE 0 END) OVER (ORDER BY date) /
            NULLIF(COUNT(*) OVER (ORDER BY date), 0),
            2
        ) as win_rate
    FROM game_results
)
SELECT json_build_object(
    'dates', json_agg(date ORDER BY date),
    'results', json_agg(result ORDER BY date),
    'cumulative_wins', json_agg(cum_wins ORDER BY date),
    'cumulative_losses', json_agg(cum_losses ORDER BY date),
    'win_rate', json_agg(win_rate ORDER BY date)
) as graph_data
FROM cumulative_stats;

WITH target_games AS (
    SELECT DISTINCT g.id, g.timestamp
    FROM games g
    JOIN teams t ON t.game_id = g.id
    JOIN teams_players tp ON tp.team_id = t.id
    JOIN players p ON p.id = tp.player_id
    WHERE g.file LIKE '%/mahendra/Windows/%' AND p.name IN ('Pineapple', 'MuNi', 'SINDIOS', 'Sajoma')
    GROUP BY g.id, g.timestamp
    HAVING COUNT(DISTINCT p.name) = 4
),
team_types AS (
    SELECT
        g.id,
        t.id as team_id,
        t.result,
        CASE WHEN COUNT(CASE WHEN p.name IN ('Pineapple', 'MuNi', 'SINDIOS') THEN 1 END) = 3 THEN 1
                 WHEN COUNT(CASE WHEN p.name = 'Sajoma' THEN 1 END) = 1 THEN 2
        END as team_type
    FROM target_games tg
    JOIN games g ON g.id = tg.id
    JOIN teams t ON t.game_id = g.id
    JOIN teams_players tp ON tp.team_id = t.id
    JOIN players p ON p.id = tp.player_id
    GROUP BY g.id, t.id, t.result
    HAVING COUNT(CASE WHEN p.name IN ('Pineapple', 'MuNi', 'SINDIOS') THEN 1 END) = 3
         OR COUNT(CASE WHEN p.name = 'Sajoma' THEN 1 END) = 1
),
team_compositions AS (
    SELECT
        tt.id,
        tt.team_id,
        tt.result,
        tt.team_type,
        STRING_AGG(
            p.name || ' (' || LEFT(tp.races_assigned, 1) || ')',
            ', '
            ORDER BY
                CASE
                    WHEN tt.team_type = 1 AND p.name IN ('Pineapple', 'MuNi', 'SINDIOS') THEN 0
                    WHEN tt.team_type = 2 AND p.name = 'Sajoma' THEN 0
                    ELSE 1
                END,
                p.name
        ) as team_info
    FROM team_types tt
    JOIN teams_players tp ON tp.team_id = tt.team_id
    JOIN players p ON p.id = tp.player_id
    GROUP BY tt.id, tt.team_id, tt.result, tt.team_type
)
SELECT
    g.timestamp,
    t1.team_info as team_1,
    t1.result as team_1_result,
    t2.team_info as team_2,
    t2.result as team_2_result,
    g.map,
    TO_CHAR((g.duration || ' seconds')::interval, 'MI:SS') as duration
FROM target_games tg
JOIN games g ON g.id = tg.id
JOIN team_compositions t1 ON t1.id = g.id AND t1.team_type = 1
JOIN team_compositions t2 ON t2.id = g.id AND t2.team_type = 2
ORDER BY g.timestamp ASC;
