package entity

import (
	"encoding/json"
	"time"
)

type Game struct {
	ID                               int             `db:"id"`
	TagIDs                           []int           `db:"tag_ids"`
	SteamID                          int             `db:"steam_id"`
	Name                             string          `db:"name"`
	StoreURLPath                     string          `db:"store_url_path"`
	IsFree                           bool            `db:"is_free"`
	Windows                          bool            `db:"windows"`
	SteamOSLinux                     bool            `db:"steamos_linux"`
	Mac                              bool            `db:"mac"`
	ReviewsFilteredCount             int             `db:"reviews_filtered_count"`
	ReviewsFilteredPercentPositive   int             `db:"reviews_filtered_percent_positive"`
	ReviewsFilteredScoreLabel        string          `db:"reviews_filtered_score_label"`
	ReviewsUnfilteredCount           int             `db:"reviews_unfiltered_count"`
	ReviewsUnfilteredPercentPositive int             `db:"reviews_unfiltered_percent_positive"`
	ReviewsUnfilteredScoreLabel      string          `db:"reviews_unfiltered_score_label"`
	Assets                           json.RawMessage `db:"assets"`
	ReleaseSteamDate                 time.Time       `db:"release_steam_date"`
	ReleaseOriginalSteamDate         time.Time       `db:"release_original_steam_date"`
}

type Developer struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

type Publisher struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

type Tag struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

type GameTag struct {
	GameID int `db:"game_id"`
	TagID  int `db:"tag_id"`
	Weight int `db:"weight"`
}

type GameDeveloper struct {
	GameID      int `db:"game_id"`
	DeveloperID int `db:"developer_id"`
}

type GamePublisher struct {
	GameID      int `db:"game_id"`
	PublisherID int `db:"publisher_id"`
}

type App struct {
	ID           int        `db:"id"`
	SteamID      int        `db:"steam_id"`
	Name         string     `db:"name"`
	LastModified time.Time  `db:"last_modified"`
	LastParsed   *time.Time `db:"last_parsed"`
}
