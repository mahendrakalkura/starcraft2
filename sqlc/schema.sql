DROP SCHEMA public CASCADE;

CREATE SCHEMA public;

CREATE TABLE files (
    id BIGSERIAL PRIMARY KEY,
    path TEXT NOT NULL UNIQUE,
    parser_version TEXT NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    status TEXT NOT NULL
);

CREATE TABLE games (
    id BIGSERIAL PRIMARY KEY,
    file_id BIGINT NOT NULL UNIQUE REFERENCES files(id) ON DELETE CASCADE,
    amm BOOLEAN NOT NULL,
    competitive BOOLEAN NOT NULL,
    duration BIGINT NOT NULL,
    fingerprint TEXT NOT NULL UNIQUE,
    map TEXT NOT NULL,
    mode TEXT NOT NULL,
    played_at TIMESTAMPTZ NOT NULL,
    version TEXT NOT NULL
);

CREATE TABLE players (
    id BIGSERIAL PRIMARY KEY,
    game_id BIGINT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    apm INT NOT NULL,
    clan TEXT NOT NULL,
    color TEXT NOT NULL,
    mmr INT NOT NULL,
    name TEXT NOT NULL,
    number INT NOT NULL,
    race_assigned TEXT NOT NULL,
    race_selected TEXT NOT NULL,
    result TEXT NOT NULL,
    team INT NOT NULL,
    UNIQUE (game_id, number)
);

CREATE TABLE messages (
    player_id BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    recipient INT NOT NULL,
    text TEXT NOT NULL,
    tick INT NOT NULL
);

CREATE INDEX messages_player_id_tick ON messages(player_id, tick);

CREATE TABLE stats (
    player_id BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    tick INT NOT NULL,
    food_made INT NOT NULL,
    food_used INT NOT NULL,
    minerals_collection_rate INT NOT NULL,
    minerals_current INT NOT NULL,
    minerals_friendly_fire_army INT NOT NULL,
    minerals_friendly_fire_economy INT NOT NULL,
    minerals_friendly_fire_technology INT NOT NULL,
    minerals_killed_army INT NOT NULL,
    minerals_killed_economy INT NOT NULL,
    minerals_killed_technology INT NOT NULL,
    minerals_lost_army INT NOT NULL,
    minerals_lost_economy INT NOT NULL,
    minerals_lost_technology INT NOT NULL,
    minerals_used_active_forces INT NOT NULL,
    minerals_used_current_army INT NOT NULL,
    minerals_used_current_economy INT NOT NULL,
    minerals_used_current_technology INT NOT NULL,
    minerals_used_in_progress_army INT NOT NULL,
    minerals_used_in_progress_economy INT NOT NULL,
    minerals_used_in_progress_technology INT NOT NULL,
    vespene_collection_rate INT NOT NULL,
    vespene_current INT NOT NULL,
    vespene_friendly_fire_army INT NOT NULL,
    vespene_friendly_fire_economy INT NOT NULL,
    vespene_friendly_fire_technology INT NOT NULL,
    vespene_killed_army INT NOT NULL,
    vespene_killed_economy INT NOT NULL,
    vespene_killed_technology INT NOT NULL,
    vespene_lost_army INT NOT NULL,
    vespene_lost_economy INT NOT NULL,
    vespene_lost_technology INT NOT NULL,
    vespene_used_active_forces INT NOT NULL,
    vespene_used_current_army INT NOT NULL,
    vespene_used_current_economy INT NOT NULL,
    vespene_used_current_technology INT NOT NULL,
    vespene_used_in_progress_army INT NOT NULL,
    vespene_used_in_progress_economy INT NOT NULL,
    vespene_used_in_progress_technology INT NOT NULL,
    workers_active_count INT NOT NULL,
    PRIMARY KEY (player_id, tick)
);

CREATE TABLE units (
    player_id BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    action TEXT NOT NULL,
    name TEXT NOT NULL,
    tick INT NOT NULL,
    x INT NOT NULL,
    y INT NOT NULL
);

CREATE INDEX units_player_id_tick ON units(player_id, tick);

CREATE VIEW matches AS
SELECT g.id, g.played_at, g.map, g.mode, g.duration, g.amm, g.competitive,
       w.names AS winners, l.names AS losers
FROM games g
LEFT JOIN LATERAL (SELECT string_agg(p.name, ', ' ORDER BY p.name) AS names FROM players p WHERE p.game_id = g.id AND p.result = 'Win') w ON true
LEFT JOIN LATERAL (SELECT string_agg(p.name, ', ' ORDER BY p.name) AS names FROM players p WHERE p.game_id = g.id AND p.result = 'Loss') l ON true;
