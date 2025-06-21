package dto

import "time"

// MatchPreview is a entry for a given match, with it's metadata and previews for each player.
type MatchPreview struct {
	Metadata *MatchPreviewMetadata
	Data     []*MatchPreviewData
}

// MatchPreviewData holds basic data from a match for a given player.
type MatchPreviewData struct {
	GameName      string `json:"game_name"`
	Tag           string `json:"tagline"`
	Region        string `json:"region"`
	Assists       int    `json:"assists"`
	Kills         int    `json:"kills"`
	Deaths        int    `json:"deaths"`
	ChampionLevel int    `json:"champion_level"`
	ChampionID    int    `json:"champion_id"`
	Items         []int  `json:"items"`
	TotalCs       int    `json:"total_cs"`
	Win           bool   `json:"win"`
	QueueID       int    `json:"queue_id"`
}

// MatchPreviewList is a map with the match ids as keys and the match data as values.
type MatchPreviewList map[string]*MatchPreview

// MatchPreviewMetadata holds a given match metadata.
type MatchPreviewMetadata struct {
	AverageElo string
	Duration   int
	Date       time.Time
	MatchId    string
	InternalId uint
	QueueId    int
}
