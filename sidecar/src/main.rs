//! sc2json: parse one .SC2Replay file into compact JSON on stdout.
//!
//! Output is a single JSON document: game metadata, players (observers excluded),
//! and per-player messages, stats, and unit events. Games with any Computer
//! participant produce {"skip": "ai"}. Errors go to stderr with exit code 1.

use std::collections::HashMap;
use std::path::PathBuf;
use std::process::ExitCode;

use chrono::DateTime;
use s2protocol::message_events::{GameEMessageRecipient, ReplayMessageEvent};
use s2protocol::tracker_events::ReplayTrackerEvent;
use serde::{Deserialize, Serialize};

const PARSER_VERSION: &str = "2";

const CONTROL_COMPUTER: u8 = 3;
const OBSERVE_PARTICIPANT: u8 = 0;

// Windows FILETIME epoch (1601-01-01) to Unix epoch (1970-01-01) in seconds.
const FILETIME_UNIX_OFFSET: i64 = 11_644_473_600;

#[derive(Serialize)]
struct Output {
    parser_version: &'static str,
    game: Game,
    players: Vec<Player>,
}

#[derive(Serialize)]
struct Skip {
    parser_version: &'static str,
    skip: &'static str,
}

#[derive(Serialize)]
struct Game {
    amm: bool,
    competitive: bool,
    duration: i64,
    map: String,
    mode: String,
    played_at: String,
    version: String,
}

#[derive(Serialize)]
struct Player {
    apm: i32,
    clan: String,
    color: String,
    messages: Vec<Message>,
    mmr: i32,
    name: String,
    number: i32,
    race_assigned: String,
    race_selected: String,
    result: String,
    stats: Vec<Stats>,
    team: i32,
    units: Vec<Unit>,
}

#[derive(Serialize)]
struct Message {
    recipient: i32,
    text: String,
    tick: i64,
}

#[derive(Serialize)]
struct Stats {
    tick: i64,
    food_made: i32,
    food_used: i32,
    minerals_collection_rate: i32,
    minerals_current: i32,
    minerals_friendly_fire_army: i32,
    minerals_friendly_fire_economy: i32,
    minerals_friendly_fire_technology: i32,
    minerals_killed_army: i32,
    minerals_killed_economy: i32,
    minerals_killed_technology: i32,
    minerals_lost_army: i32,
    minerals_lost_economy: i32,
    minerals_lost_technology: i32,
    minerals_used_active_forces: i32,
    minerals_used_current_army: i32,
    minerals_used_current_economy: i32,
    minerals_used_current_technology: i32,
    minerals_used_in_progress_army: i32,
    minerals_used_in_progress_economy: i32,
    minerals_used_in_progress_technology: i32,
    vespene_collection_rate: i32,
    vespene_current: i32,
    vespene_friendly_fire_army: i32,
    vespene_friendly_fire_economy: i32,
    vespene_friendly_fire_technology: i32,
    vespene_killed_army: i32,
    vespene_killed_economy: i32,
    vespene_killed_technology: i32,
    vespene_lost_army: i32,
    vespene_lost_economy: i32,
    vespene_lost_technology: i32,
    vespene_used_active_forces: i32,
    vespene_used_current_army: i32,
    vespene_used_current_economy: i32,
    vespene_used_current_technology: i32,
    vespene_used_in_progress_army: i32,
    vespene_used_in_progress_economy: i32,
    vespene_used_in_progress_technology: i32,
    workers_active_count: i32,
}

#[derive(Serialize)]
struct Unit {
    action: &'static str,
    name: String,
    tick: i64,
    x: i32,
    y: i32,
}

#[derive(Deserialize)]
struct Metadata {
    #[serde(rename = "Duration", default)]
    duration: i64,
    #[serde(rename = "GameVersion", default)]
    game_version: String,
    #[serde(rename = "Players", default)]
    players: Vec<MetadataPlayer>,
}

#[derive(Deserialize)]
struct MetadataPlayer {
    #[serde(rename = "APM", default)]
    apm: f64,
    #[serde(rename = "AssignedRace", default)]
    assigned_race: String,
    #[serde(rename = "PlayerID")]
    player_id: u8,
    #[serde(rename = "Result", default)]
    result: String,
    #[serde(rename = "SelectedRace", default)]
    selected_race: String,
}

fn main() -> ExitCode {
    let arguments: Vec<String> = std::env::args().collect();

    if arguments.len() == 2 && arguments[1] == "--version" {
        println!("{}", PARSER_VERSION);
        return ExitCode::SUCCESS;
    }

    if arguments.len() != 2 {
        eprintln!("usage: sc2json <replay.SC2Replay> | sc2json --version");
        return ExitCode::FAILURE;
    }

    match run(&arguments[1]) {
        Ok(json) => {
            println!("{}", json);
            ExitCode::SUCCESS
        }
        Err(error) => {
            eprintln!("sc2json: {}: {}", arguments[1], error);
            ExitCode::FAILURE
        }
    }
}

fn run(path: &str) -> Result<String, Box<dyn std::error::Error>> {
    let file_contents = s2protocol::read_file(&PathBuf::from(path))?;
    let (_input, mpq) = s2protocol::parser::parse(&file_contents)
        .map_err(|error| format!("mpq parse: {:?}", error))?;

    let details = s2protocol::read_details(path, &mpq, &file_contents)?;

    let participants: Vec<&s2protocol::details::PlayerDetails> = details
        .player_list
        .iter()
        .filter(|player| player.observe == OBSERVE_PARTICIPANT)
        .collect();

    if participants.iter().any(|player| player.control == CONTROL_COMPUTER) {
        let skip = Skip {
            parser_version: PARSER_VERSION,
            skip: "ai",
        };
        return Ok(serde_json::to_string(&skip)?);
    }

    let (_tail, metadata_bytes) = mpq
        .read_mpq_file_sector("replay.gamemetadata.json", false, &file_contents)
        .map_err(|error| format!("gamemetadata read: {:?}", error))?;
    let metadata: Metadata = serde_json::from_slice(&metadata_bytes)?;

    let init_data = s2protocol::read_init_data(path, &mpq, &file_contents)?;
    let message_events = s2protocol::read_message_events(path, &mpq, &file_contents)?;
    let tracker_events = s2protocol::read_tracker_events(path, &mpq, &file_contents)?;

    // Tracker player_id -> (user_id, slot_id) from PlayerSetup events.
    let mut setup_user_ids: HashMap<u8, u32> = HashMap::new();
    let mut setup_slot_ids: HashMap<u8, u32> = HashMap::new();
    for tracker_event in &tracker_events {
        if let ReplayTrackerEvent::PlayerSetup(setup) = &tracker_event.event {
            if let Some(user_id) = setup.user_id {
                setup_user_ids.insert(setup.player_id, user_id);
            }
            if let Some(slot_id) = setup.slot_id {
                setup_slot_ids.insert(setup.player_id, slot_id);
            }
        }
    }

    // Tracker player_id -> details player, via working_set_slot_id, falling back to list order.
    let mut details_by_number: HashMap<u8, &s2protocol::details::PlayerDetails> = HashMap::new();
    for (index, player) in participants.iter().enumerate() {
        let number = setup_slot_ids
            .iter()
            .find(|(_, slot_id)| player.working_set_slot_id == Some(**slot_id as u8))
            .map(|(player_id, _)| *player_id)
            .unwrap_or(index as u8 + 1);
        details_by_number.insert(number, player);
    }

    let user_initial_data = &init_data.sync_lobby_state.user_initial_data;
    let game_options = &init_data.sync_lobby_state.game_description.game_options;

    let mut players: Vec<Player> = Vec::new();
    for (number, detail) in &details_by_number {
        let user_id = setup_user_ids.get(number).copied();
        let user = user_id.and_then(|id| user_initial_data.get(id as usize));

        let metadata_player = metadata
            .players
            .iter()
            .find(|player| player.player_id == *number);

        let (clan, name) = match user {
            Some(user) if !clean(&user.name).is_empty() => (
                clean(&user.clan_tag.clone().unwrap_or_default()),
                clean(&user.name),
            ),
            _ => split_clan(&detail.name),
        };

        let result = match metadata_player {
            Some(player) if !player.result.is_empty() => normalize_result(&player.result),
            _ => detail.result.clone(),
        };

        let player = Player {
            apm: metadata_player.map(|player| player.apm.round() as i32).unwrap_or(0),
            clan,
            color: format!(
                "#{:02X}{:02X}{:02X}{:02X}",
                detail.color.r, detail.color.g, detail.color.b, detail.color.a
            ),
            messages: Vec::new(),
            mmr: user.and_then(|user| user.scaled_rating).unwrap_or(0),
            name,
            number: *number as i32,
            race_assigned: metadata_player
                .map(|player| normalize_race(&player.assigned_race))
                .unwrap_or_else(|| detail.race.clone()),
            race_selected: metadata_player
                .map(|player| normalize_race(&player.selected_race))
                .unwrap_or_else(|| detail.race.clone()),
            result,
            stats: Vec::new(),
            team: detail.team_id as i32 + 1,
            units: Vec::new(),
        };
        players.push(player);
    }
    players.sort_by_key(|player| player.number);

    // Index into players by tracker player_id.
    let mut player_index: HashMap<i32, usize> = HashMap::new();
    for (index, player) in players.iter().enumerate() {
        player_index.insert(player.number, index);
    }

    // Lobby user_id -> tracker player_id, for chat attribution.
    let mut user_to_number: HashMap<i64, i32> = HashMap::new();
    for (number, user_id) in &setup_user_ids {
        user_to_number.insert(*user_id as i64, *number as i32);
    }

    let mut tick: i64 = 0;
    for message_event in &message_events {
        tick += message_event.delta;
        let ReplayMessageEvent::EChat(chat) = &message_event.event;
        let Some(number) = user_to_number.get(&message_event.user_id) else {
            continue; // observer chat
        };
        let Some(index) = player_index.get(number) else {
            continue;
        };
        let message = Message {
            recipient: recipient_code(&chat.m_recipient),
            text: chat.m_string.clone(),
            tick,
        };
        players[*index].messages.push(message);
    }

    // Unit names by (tag_index, tag_recycle), from born and under-construction events.
    let mut unit_names: HashMap<(u32, u32), String> = HashMap::new();
    for tracker_event in &tracker_events {
        match &tracker_event.event {
            ReplayTrackerEvent::UnitBorn(born) => {
                unit_names.insert(
                    (born.unit_tag_index, born.unit_tag_recycle),
                    born.unit_type_name.clone(),
                );
            }
            ReplayTrackerEvent::UnitInit(init) => {
                unit_names.insert(
                    (init.unit_tag_index, init.unit_tag_recycle),
                    init.unit_type_name.clone(),
                );
            }
            _ => {}
        }
    }

    let mut tick: i64 = 0;
    for tracker_event in &tracker_events {
        tick += tracker_event.delta as i64;
        match &tracker_event.event {
            ReplayTrackerEvent::PlayerStats(event) => {
                let Some(index) = player_index.get(&(event.player_id as i32)) else {
                    continue;
                };
                let stats = &mut players[*index].stats;
                // The game-end snapshot reuses the last periodic tick; keep the later one.
                if stats.last().is_some_and(|last| last.tick == tick) {
                    stats.pop();
                }
                stats.push(to_stats(tick, &event.stats));
            }
            ReplayTrackerEvent::UnitBorn(event) => {
                let Some(index) = player_index.get(&(event.control_player_id as i32)) else {
                    continue; // neutral or unowned
                };
                if event.unit_type_name.contains("Dummy") {
                    continue; // engine helper units (e.g. InvisibleTargetDummy)
                }
                let unit = Unit {
                    action: "born",
                    name: event.unit_type_name.clone(),
                    tick,
                    x: event.x as i32,
                    y: event.y as i32,
                };
                players[*index].units.push(unit);
            }
            ReplayTrackerEvent::UnitInit(event) => {
                let Some(index) = player_index.get(&(event.control_player_id as i32)) else {
                    continue; // neutral or unowned
                };
                if event.unit_type_name.contains("Dummy") {
                    continue; // engine helper units (e.g. InvisibleTargetDummy)
                }
                let unit = Unit {
                    action: "init",
                    name: event.unit_type_name.clone(),
                    tick,
                    x: event.x as i32,
                    y: event.y as i32,
                };
                players[*index].units.push(unit);
            }
            ReplayTrackerEvent::UnitDied(event) => {
                let Some(killer) = event.killer_player_id else {
                    continue; // killerless deaths are dropped
                };
                let Some(index) = player_index.get(&(killer as i32)) else {
                    continue;
                };
                let Some(name) = unit_names.get(&(event.unit_tag_index, event.unit_tag_recycle))
                else {
                    continue;
                };
                if name.contains("Dummy") {
                    continue; // engine helper units (e.g. InvisibleTargetDummy)
                }
                let unit = Unit {
                    action: "killed",
                    name: name.clone(),
                    tick,
                    x: event.x as i32,
                    y: event.y as i32,
                };
                players[*index].units.push(unit);
            }
            _ => {}
        }
    }

    let mut team_sizes: HashMap<i32, i32> = HashMap::new();
    for player in &players {
        *team_sizes.entry(player.team).or_insert(0) += 1;
    }
    let mut sizes: Vec<i32> = team_sizes.values().copied().collect();
    sizes.sort_unstable();
    let mode = sizes
        .iter()
        .map(|size| size.to_string())
        .collect::<Vec<String>>()
        .join("v");

    let played_at_seconds = details.time_utc / 10_000_000 - FILETIME_UNIX_OFFSET;
    let played_at = DateTime::from_timestamp(played_at_seconds, 0)
        .ok_or("invalid time_utc")?
        .to_rfc3339();

    let output = Output {
        parser_version: PARSER_VERSION,
        game: Game {
            amm: game_options.amm,
            competitive: game_options.competitive,
            duration: metadata.duration,
            map: details.title.clone(),
            mode,
            played_at,
            version: metadata.game_version,
        },
        players,
    };

    Ok(serde_json::to_string(&output)?)
}

// Init data strings can carry stray NUL bytes; strip them and surrounding whitespace.
fn clean(value: &str) -> String {
    String::from(value.replace('\0', "").trim())
}

fn normalize_race(race: &str) -> String {
    match race {
        "Prot" => String::from("Protoss"),
        "Rand" => String::from("Random"),
        "Terr" => String::from("Terran"),
        other => String::from(other),
    }
}

fn normalize_result(result: &str) -> String {
    match result {
        "Loss" | "Tie" | "Undecided" | "Win" => String::from(result),
        other => String::from(other),
    }
}

fn recipient_code(recipient: &GameEMessageRecipient) -> i32 {
    match recipient {
        GameEMessageRecipient::EAll => 0,
        GameEMessageRecipient::EAllies => 2,
        GameEMessageRecipient::EIndividual => 1,
        GameEMessageRecipient::EBattlenet => 3,
        GameEMessageRecipient::EObservers => 4,
    }
}

// Details names embed the clan as "&lt;CLAN&gt;<sp/>Name"; split it off.
fn split_clan(name: &str) -> (String, String) {
    if let Some(rest) = name.strip_prefix("&lt;")
        && let Some((clan, player)) = rest.split_once("&gt;")
    {
        let player = player.strip_prefix("<sp/>").unwrap_or(player);
        return (String::from(clan), String::from(player));
    }
    (String::new(), String::from(name))
}

fn to_stats(tick: i64, stats: &s2protocol::tracker_events::PlayerStats) -> Stats {
    Stats {
        tick,
        food_made: stats.food_made,
        food_used: stats.food_used,
        minerals_collection_rate: stats.minerals_collection_rate,
        minerals_current: stats.minerals_current,
        minerals_friendly_fire_army: stats.minerals_friendly_fire_army,
        minerals_friendly_fire_economy: stats.minerals_friendly_fire_economy,
        minerals_friendly_fire_technology: stats.minerals_friendly_fire_technology,
        minerals_killed_army: stats.minerals_killed_army,
        minerals_killed_economy: stats.minerals_killed_economy,
        minerals_killed_technology: stats.minerals_killed_technology,
        minerals_lost_army: stats.minerals_lost_army,
        minerals_lost_economy: stats.minerals_lost_economy,
        minerals_lost_technology: stats.minerals_lost_technology,
        minerals_used_active_forces: stats.minerals_used_active_forces,
        minerals_used_current_army: stats.minerals_used_current_army,
        minerals_used_current_economy: stats.minerals_used_current_economy,
        minerals_used_current_technology: stats.minerals_used_current_technology,
        minerals_used_in_progress_army: stats.minerals_used_in_progress_army,
        minerals_used_in_progress_economy: stats.minerals_used_in_progress_economy,
        minerals_used_in_progress_technology: stats.minerals_used_in_progress_technology,
        vespene_collection_rate: stats.vespene_collection_rate,
        vespene_current: stats.vespene_current,
        vespene_friendly_fire_army: stats.vespene_friendly_fire_army,
        vespene_friendly_fire_economy: stats.vespene_friendly_fire_economy,
        vespene_friendly_fire_technology: stats.vespene_friendly_fire_technology,
        vespene_killed_army: stats.vespene_killed_army,
        vespene_killed_economy: stats.vespene_killed_economy,
        vespene_killed_technology: stats.vespene_killed_technology,
        vespene_lost_army: stats.vespene_lost_army,
        vespene_lost_economy: stats.vespene_lost_economy,
        vespene_lost_technology: stats.vespene_lost_technology,
        vespene_used_active_forces: stats.vespene_used_active_forces,
        vespene_used_current_army: stats.vespene_used_current_army,
        vespene_used_current_economy: stats.vespene_used_current_economy,
        vespene_used_current_technology: stats.vespene_used_current_technology,
        vespene_used_in_progress_army: stats.vespene_used_in_progress_army,
        vespene_used_in_progress_economy: stats.vespene_used_in_progress_economy,
        vespene_used_in_progress_technology: stats.vespene_used_in_progress_technology,
        workers_active_count: stats.workers_active_count,
    }
}
