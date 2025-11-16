package dto

import (
	"time"

	"gorm.io/datatypes"
)

// MatchPreview is a entry for a given match, with it's metadata and previews for each player.
type MatchPreview struct {
	Metadata *MatchPreviewMetadata `json:"metadata"`
	Data     []*MatchPreviewData   `json:"data"`
}

// MatchPreviewData holds basic data from a match for a given player.
type MatchPreviewData struct {
	PlayerId      uint   `json:"playerId"`
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
	ParticipantId int    `json:"participantId"`
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

func NewMatchPreviewList() MatchPreviewList {
	return make(MatchPreviewList)
}

func (mpl MatchPreviewList) AddMatch(matchId string, preview *MatchPreview) {
	mpl[matchId] = preview
}

// FullMatchData consists in all available formatted data for the matches.
type FullMatchData struct {
	Metadata             *MatchPreviewMetadata `json:"metadata"`
	ParticipantsPreviews []*MatchPreviewData   `json:"participants"`
	ParticipantFrames    ParticipantFrameList  `json:"participant_frames"`
	Events               []MatchEvents         `json:"events"`
}

// ParticipantFrameList is the map of frames for each participant in a given match.
type ParticipantFrameList map[int][]ParticipantFrame

// NewParticipantFrameList allocate and return a  participant frame list.
func NewParticipantFrameList() ParticipantFrameList {
	return make(ParticipantFrameList)
}

func (pfl ParticipantFrameList) AddFrame(frame ParticipantFrame) {
	pfl[frame.ParticipantID] = append(pfl[frame.ParticipantID], frame)
}

type ParticipantFrame struct {
	CurrentGold                   int `json:"currentGold"`
	FrameIndex                    int `json:"frameIndex"`
	JungleMinionsKilled           int `json:"jungleMinionsKilled"`
	Level                         int `json:"level"`
	MagicDamageDone               int `json:"magicDamageDone"`
	MagicDamageDoneToChampions    int `json:"magicDamageDoneToChampions"`
	MagicDamageTaken              int `json:"magicDamageTaken"`
	MatchStatID                   int `json:"matchStatId"`
	MinionsKilled                 int `json:"minionsKilled"`
	ParticipantID                 int `json:"participantId"`
	PhysicalDamageDone            int `json:"physicalDamageDone"`
	PhysicalDamageDoneToChampions int `json:"physicalDamageDoneToChampions"`
	PhysicalDamageTaken           int `json:"physicalDamageTaken"`
	TotalDamageDone               int `json:"totalDamageDone"`
	TotalDamageDoneToChampions    int `json:"totalDamageDoneToChampions"`
	TotalDamageTaken              int `json:"totalDamageTaken"`
	TotalGold                     int `json:"totalGold"`
	TrueDamageDone                int `json:"trueDamageDone"`
	TrueDamageDoneToChampions     int `json:"trueDamageDoneToChampions"`
	TrueDamageTaken               int `json:"trueDamageTaken"`
	XP                            int `json:"xp"`
}

// MatchEvents holds data regarding a given event in a match.
type MatchEvents struct {
	Timestamp     int64          `json:"timestamp"`
	EventType     string         `json:"eventType"`
	ParticipantId *int           `json:"participantId"`
	Data          datatypes.JSON `json:"data"`
}
