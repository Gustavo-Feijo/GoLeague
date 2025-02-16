package match_fetcher

// Match information.
type MatchInfo struct {
	EndOfGameResult string        `json:"endOfGameResult"`
	GameCreation    RiotTime      `json:"gameCreation"`
	GameDuration    int           `json:"gameDuration"`
	GameVersion     string        `json:"gameVersion"`
	Participants    []MatchPlayer `json:"participants"`
	QueueId         int           `json:"queueId"`
	Teams           []TeamInfo    `json:"teams"`
}

// Player results.
type MatchPlayer struct {
	Assists       int16      `json:"assists"`
	AssistMePing  int16      `json:"assistMePing"`
	BaronKills    int16      `json:"baronKills"`
	BasicPings    int16      `json:"basicPings"`
	ChampionLevel int8       `json:"champLevel"`
	ChampionId    int16      `json:"championId"`
	Challenges    Challenges `json:"challenges"`
	Deaths        int16      `json:"deaths"`
	Kills         int16      `json:"kills"`
}

// Challenges of the player for this match.
// Some entries like KDA and GoldPerMinute can be fetched here.
// However we can calculate at runtime without storing it.
type Challenges struct {
	AbilityUses        int16 `json:"abilityUses"`
	ControlWardsPlaced int16 `json:"controlWardsPlaced"`
	SkillshotsDodged   int16 `json:"skillshotsDodged"`
}

// Team information.
type TeamInfo struct {
	Bans   []Ban `json:"bans"`
	TeamId int16 `json:"teamId"`
	Win    bool  `json:"win"`
}

// Ban information.
type Ban struct {
	ChampionId int16 `json:"championId"`
	PickTurn   int8  `json:"pickTurn"`
}
