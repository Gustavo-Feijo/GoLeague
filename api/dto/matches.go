package dto

import (
	"time"
)

// MatchPreview is a entry for a given match, with it's metadata and previews for each player.
type MatchPreview struct {
	Metadata *MatchPreviewMetadata `json:"metadata"`
	Data     []*MatchPreviewData   `json:"data"`
}

// MatchPreviewData holds basic data from a match for a given player.
type MatchPreviewData struct {
	GameName      string `json:"gameName"`
	Tag           string `json:"tagLine"`
	Region        string `json:"region"`
	Assists       int    `json:"assists"`
	Kills         int    `json:"kills"`
	Deaths        int    `json:"deaths"`
	ChampionLevel int    `json:"championLevel"`
	ChampionID    int    `json:"championId"`
	TeamId        int    `json:"teamId"`
	Items         []int  `json:"items"`
	TotalCs       int    `json:"totalCs"`
	Win           bool   `json:"win"`
	QueueID       int    `json:"queueId"`
}

// MatchPreviewList is a map with the match ids as keys and the match data as values.
type MatchPreviewList map[string]*MatchPreview

// MatchPreviewMetadata holds a given match metadata.
type MatchPreviewMetadata struct {
	AverageElo   string    `json:"averageElo"`
	Duration     int       `json:"duration"`
	Date         time.Time `json:"date"`
	MatchId      string    `json:"matchId"`
	InternalId   uint      `json:"internalId"`
	QueueId      int       `json:"queueId"`
	WinnerTeamId int       `json:"winnerTeamId"`
}
