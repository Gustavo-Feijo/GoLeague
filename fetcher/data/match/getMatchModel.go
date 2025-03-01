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
	Assists                        uint16     `json:"assists"`
	AllInPings                     uint16     `json:"allInPings"`
	AssistMePing                   uint16     `json:"assistMePing"`
	BaronKills                     uint16     `json:"baronKills"`
	BasicPings                     uint16     `json:"basicPings"`
	ChampionLevel                  uint8      `json:"champLevel"`
	ChampionId                     uint16     `json:"championId"`
	Challenges                     Challenges `json:"challenges"`
	CommandPings                   uint16     `json:"commandPings"`
	DangerPings                    uint16     `json:"dangerPings"`
	Deaths                         uint16     `json:"deaths"`
	EnemyMissingPings              uint16     `json:"enemyMissingPings"`
	EnemyVisionPings               uint16     `json:"enemyVisionPings"`
	GameEndedInEarlySurrender      bool       `json:"gameEndedInEarlySurrender"`
	GameEndedInSurrender           bool       `json:"gameEndedInSurrender"`
	GetBackPings                   uint16     `json:"getBackPings"`
	GoldEarned                     uint32     `json:"GoldEarned"`
	GoldSpent                      uint32     `json:"goldSpent"`
	HoldPings                      uint16     `json:"holdPings"`
	Item0                          uint16     `json:"item0"`
	Item1                          uint16     `json:"item1"`
	Item2                          uint16     `json:"item2"`
	Item3                          uint16     `json:"item3"`
	Item4                          uint16     `json:"item4"`
	Item5                          uint16     `json:"item5"`
	Kills                          uint16     `json:"kills"`
	MagicDamageDealtToChampions    uint32     `json:"magicDamageDealtToChampions"`
	MagicDamageTaken               uint32     `json:"magicDamageTaken"`
	NeedVisionPings                uint16     `json:"needVisionPings"`
	NeutralMinionsKilled           uint16     `json:"neutralMinionsKilled"`
	OnMyWayPings                   uint16     `json:"onMyWayPings"`
	PhysicalDamageDealtToChampions uint32     `json:"physicalDamageDealtToChampions"`
	PhysicalDamageTaken            uint32     `json:"physicalDamageTaken"`
	ProfileIcon                    uint16     `json:"profileIcon"`
	PushPings                      uint16     `json:"pushPings"`
	Puuid                          string     `json:"puuid"`
	RetreatPings                   uint16     `json:"retreatPings"`
	RiotIdGameName                 string     `json:"riotIdGameName"`
	RiotIdTagline                  string     `json:"riotIdTagline"`
	SummonerId                     string     `json:"summonerId"`
	SummonerLevel                  uint16     `json:"summonerLevel"`
	LongestTimeSpentLiving         uint16     `json:"longestTimeSpentLiving"`
	MagicDamageDealt               uint32     `json:"magicDamageDealt"`
	TeamId                         uint16     `json:"teamId"`
	TeamPosition                   string     `json:"teamPosition"`
	TimeCCingOthers                uint16     `json:"timeCCingOthers"`
	TotalDamageDealtToChampions    uint32     `json:"totalDamageDealtToChampions"`
	TotalMinionsKilled             uint16     `json:"totalMinionsKilled"`
	TotalTimeSpentDead             uint16     `json:"totalTimeSpentDead"`
	TrueDamageDealtToChampions     uint32     `json:"trueDamageDealtToChampions"`
	VisionClearedPings             uint16     `json:"visionClearedPings"`
	VisionScore                    uint16     `json:"visionScore"`
	WardsKilled                    uint16     `json:"wardsKilled"`
	WardsPlaced                    uint16     `json:"wardsPlaced"`
	Win                            uint16     `json:"win"`
}

// Challenges of the player for this match.
// Some entries like KDA and GoldPerMinute can be fetched here.
// However we can calculate at runtime without storing it.
type Challenges struct {
	AbilityUses        uint16 `json:"abilityUses"`
	ControlWardsPlaced uint16 `json:"controlWardsPlaced"`
	SkillshotsDodged   uint16 `json:"skillshotsDodged"`
	SkillshotsHit      uint16 `json:"skillshotsHit"`
}

// Team information.
type TeamInfo struct {
	Bans   []Ban  `json:"bans"`
	TeamId uint16 `json:"teamId"`
	Win    bool   `json:"win"`
}

// Ban information.
type Ban struct {
	ChampionId int16 `json:"championId"`
	PickTurn   uint8 `json:"pickTurn"`
}
