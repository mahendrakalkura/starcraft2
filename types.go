package main

import "time"

type Game struct {
	Amm         bool      `json:"amm"`
	Competitive bool      `json:"competitive"`
	Duration    int64     `json:"duration"`
	Map         string    `json:"map"`
	Mode        string    `json:"mode"`
	PlayedAt    time.Time `json:"played_at"`
	Version     string    `json:"version"`
}

type Message struct {
	Recipient int32  `json:"recipient"`
	Text      string `json:"text"`
	Tick      int32  `json:"tick"`
}

type Player struct {
	Apm          int32     `json:"apm"`
	Clan         string    `json:"clan"`
	Color        string    `json:"color"`
	Messages     []Message `json:"messages"`
	Mmr          int32     `json:"mmr"`
	Name         string    `json:"name"`
	Number       int32     `json:"number"`
	RaceAssigned string    `json:"race_assigned"`
	RaceSelected string    `json:"race_selected"`
	Result       string    `json:"result"`
	Stats        []Stats   `json:"stats"`
	Team         int32     `json:"team"`
	Units        []Unit    `json:"units"`
}

type Replay struct {
	ParserVersion string   `json:"parser_version"`
	Skip          string   `json:"skip"`
	Game          Game     `json:"game"`
	Players       []Player `json:"players"`
}

type Stats struct {
	Tick                             int32 `json:"tick"`
	FoodMade                         int32 `json:"food_made"`
	FoodUsed                         int32 `json:"food_used"`
	MineralsCollectionRate           int32 `json:"minerals_collection_rate"`
	MineralsCurrent                  int32 `json:"minerals_current"`
	MineralsFriendlyFireArmy         int32 `json:"minerals_friendly_fire_army"`
	MineralsFriendlyFireEconomy      int32 `json:"minerals_friendly_fire_economy"`
	MineralsFriendlyFireTechnology   int32 `json:"minerals_friendly_fire_technology"`
	MineralsKilledArmy               int32 `json:"minerals_killed_army"`
	MineralsKilledEconomy            int32 `json:"minerals_killed_economy"`
	MineralsKilledTechnology         int32 `json:"minerals_killed_technology"`
	MineralsLostArmy                 int32 `json:"minerals_lost_army"`
	MineralsLostEconomy              int32 `json:"minerals_lost_economy"`
	MineralsLostTechnology           int32 `json:"minerals_lost_technology"`
	MineralsUsedActiveForces         int32 `json:"minerals_used_active_forces"`
	MineralsUsedCurrentArmy          int32 `json:"minerals_used_current_army"`
	MineralsUsedCurrentEconomy       int32 `json:"minerals_used_current_economy"`
	MineralsUsedCurrentTechnology    int32 `json:"minerals_used_current_technology"`
	MineralsUsedInProgressArmy       int32 `json:"minerals_used_in_progress_army"`
	MineralsUsedInProgressEconomy    int32 `json:"minerals_used_in_progress_economy"`
	MineralsUsedInProgressTechnology int32 `json:"minerals_used_in_progress_technology"`
	VespeneCollectionRate            int32 `json:"vespene_collection_rate"`
	VespeneCurrent                   int32 `json:"vespene_current"`
	VespeneFriendlyFireArmy          int32 `json:"vespene_friendly_fire_army"`
	VespeneFriendlyFireEconomy       int32 `json:"vespene_friendly_fire_economy"`
	VespeneFriendlyFireTechnology    int32 `json:"vespene_friendly_fire_technology"`
	VespeneKilledArmy                int32 `json:"vespene_killed_army"`
	VespeneKilledEconomy             int32 `json:"vespene_killed_economy"`
	VespeneKilledTechnology          int32 `json:"vespene_killed_technology"`
	VespeneLostArmy                  int32 `json:"vespene_lost_army"`
	VespeneLostEconomy               int32 `json:"vespene_lost_economy"`
	VespeneLostTechnology            int32 `json:"vespene_lost_technology"`
	VespeneUsedActiveForces          int32 `json:"vespene_used_active_forces"`
	VespeneUsedCurrentArmy           int32 `json:"vespene_used_current_army"`
	VespeneUsedCurrentEconomy        int32 `json:"vespene_used_current_economy"`
	VespeneUsedCurrentTechnology     int32 `json:"vespene_used_current_technology"`
	VespeneUsedInProgressArmy        int32 `json:"vespene_used_in_progress_army"`
	VespeneUsedInProgressEconomy     int32 `json:"vespene_used_in_progress_economy"`
	VespeneUsedInProgressTechnology  int32 `json:"vespene_used_in_progress_technology"`
	WorkersActiveCount               int32 `json:"workers_active_count"`
}

type Unit struct {
	Action string `json:"action"`
	Name   string `json:"name"`
	Tick   int32  `json:"tick"`
	X      int32  `json:"x"`
	Y      int32  `json:"y"`
}
