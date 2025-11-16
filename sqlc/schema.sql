DROP SCHEMA public CASCADE;

CREATE SCHEMA public;

CREATE TABLE games (
    id BIGSERIAL PRIMARY KEY,
    file TEXT NOT NULL UNIQUE,
    duration BIGINT NOT NULL,
    map TEXT NOT NULL,
    mode TEXT NOT NULL,
    timestamp TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    type TEXT NOT NULL,
    CONSTRAINT games_file UNIQUE (file)
);

CREATE TABLE teams (
    id BIGSERIAL PRIMARY KEY,
    game_id BIGINT NOT NULL REFERENCES games(id),
    number BIGINT NOT NULL,
    result TEXT NOT NULL,
    CONSTRAINT teams_game_id_number UNIQUE (game_id, number)
);

CREATE TABLE players (
    id BIGSERIAL PRIMARY KEY,
    team_id BIGINT NOT NULL REFERENCES teams(id),
    number BIGINT NOT NULL,
    apm BIGINT NOT NULL,
    color TEXT NOT NULL,
    control TEXT NOT NULL,
    mmr BIGINT NOT NULL,
    name TEXT NOT NULL,
    observe TEXT NOT NULL,
    races_assigned TEXT NOT NULL,
    races_selected TEXT NOT NULL,
    CONSTRAINT players_team_id_name UNIQUE (team_id, number)
);

CREATE TABLE messages (
    id BIGSERIAL PRIMARY KEY,
    player_id BIGINT NOT NULL REFERENCES players(id),
    time BIGINT NOT NULL,
    recipient_id BIGINT NOT NULL,
    string TEXT NOT NULL
);

CREATE TABLE stats (
    id BIGSERIAL PRIMARY KEY,
    player_id BIGINT NOT NULL REFERENCES players(id),
    time BIGINT NOT NULL,
    food_made BIGINT NOT NULL,
    food_used BIGINT NOT NULL,
    minerals_collection_rate BIGINT NOT NULL,
    minerals_current BIGINT NOT NULL,
    minerals_friendly_fire_army BIGINT NOT NULL,
    minerals_friendly_fire_economy BIGINT NOT NULL,
    minerals_friendly_fire_technology BIGINT NOT NULL,
    minerals_killed_army BIGINT NOT NULL,
    minerals_killed_economy BIGINT NOT NULL,
    minerals_killed_technology BIGINT NOT NULL,
    minerals_lost_army BIGINT NOT NULL,
    minerals_lost_economy BIGINT NOT NULL,
    minerals_lost_technology BIGINT NOT NULL,
    minerals_used_active_forces BIGINT NOT NULL,
    minerals_used_current_army BIGINT NOT NULL,
    minerals_used_current_economy BIGINT NOT NULL,
    minerals_used_current_technology BIGINT NOT NULL,
    minerals_used_in_progress_army BIGINT NOT NULL,
    minerals_used_in_progress_economy BIGINT NOT NULL,
    minerals_used_in_progress_technology BIGINT NOT NULL,
    vespene_collection_rate BIGINT NOT NULL,
    vespene_current BIGINT NOT NULL,
    vespene_friendly_fire_army BIGINT NOT NULL,
    vespene_friendly_fire_economy BIGINT NOT NULL,
    vespene_friendly_fire_technology BIGINT NOT NULL,
    vespene_killed_army BIGINT NOT NULL,
    vespene_killed_economy BIGINT NOT NULL,
    vespene_killed_technology BIGINT NOT NULL,
    vespene_lost_army BIGINT NOT NULL,
    vespene_lost_economy BIGINT NOT NULL,
    vespene_lost_technology BIGINT NOT NULL,
    vespene_used_active_forces BIGINT NOT NULL,
    vespene_used_current_army BIGINT NOT NULL,
    vespene_used_current_economy BIGINT NOT NULL,
    vespene_used_current_technology BIGINT NOT NULL,
    vespene_used_in_progress_army BIGINT NOT NULL,
    vespene_used_in_progress_economy BIGINT NOT NULL,
    vespene_used_in_progress_technology BIGINT NOT NULL,
    workers_active_count BIGINT NOT NULL
);

CREATE TABLE units (
    id BIGSERIAL PRIMARY KEY,
    player_id BIGINT NOT NULL REFERENCES players(id),
    time BIGINT NOT NULL,
    action TEXT NOT NULL,
    name TEXT NOT NULL,
    x BIGINT NOT NULL,
    y BIGINT NOT NULL
);
