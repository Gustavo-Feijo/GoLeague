package matchfetcher

// MatchInfo contains the basic match metadata.
type MatchInfo struct {
	EndOfGameResult string        `json:"endOfGameResult"`
	GameCreation    RiotTime      `json:"gameCreation"`
	GameDuration    int           `json:"gameDuration"`
	GameMode        string        `json:"gameMode"`
	GameVersion     string        `json:"gameVersion"`
	Participants    []MatchPlayer `json:"participants"`
	PlatformId      string        `json:"platformId"`
	QueueId         int           `json:"queueId"`
	Teams           []TeamInfo    `json:"teams"`
}

// MatchPlayer contains the stats and information about a given player in a Match.
type MatchPlayer struct {
	Assists                        int        `json:"assists"`
	AllInPings                     int        `json:"allInPings"`
	AssistMePing                   int        `json:"assistMePing"`
	BaronKills                     int        `json:"baronKills"`
	BasicPings                     int        `json:"basicPings"`
	ChampionLevel                  int        `json:"champLevel"`
	ChampionId                     int        `json:"championId" gorm:"index_champion_position"`
	Challenges                     Challenges `json:"challenges" gorm:"embedded"`
	CommandPings                   int        `json:"commandPings"`
	DangerPings                    int        `json:"dangerPings"`
	Deaths                         int        `json:"deaths"`
	EnemyMissingPings              int        `json:"enemyMissingPings"`
	EnemyVisionPings               int        `json:"enemyVisionPings"`
	GameEndedInEarlySurrender      bool       `json:"gameEndedInEarlySurrender" gorm:"-"`
	GameEndedInSurrender           bool       `json:"gameEndedInSurrender" gorm:"-"`
	GetBackPings                   int        `json:"getBackPings"`
	GoldEarned                     int        `json:"goldEarned"`
	GoldSpent                      int        `json:"goldSpent"`
	HoldPings                      int        `json:"holdPings"`
	Item0                          int        `json:"item0"`
	Item1                          int        `json:"item1"`
	Item2                          int        `json:"item2"`
	Item3                          int        `json:"item3"`
	Item4                          int        `json:"item4"`
	Item5                          int        `json:"item5"`
	Kills                          int        `json:"kills"`
	MagicDamageDealtToChampions    int        `json:"magicDamageDealtToChampions"`
	MagicDamageTaken               int        `json:"magicDamageTaken"`
	NeedVisionPings                int        `json:"needVisionPings"`
	NeutralMinionsKilled           int        `json:"neutralMinionsKilled"`
	OnMyWayPings                   int        `json:"onMyWayPings"`
	PhysicalDamageDealtToChampions int        `json:"physicalDamageDealtToChampions"`
	PhysicalDamageTaken            int        `json:"physicalDamageTaken"`
	ProfileIcon                    int        `json:"profileIcon" gorm:"-"`
	PushPings                      int        `json:"pushPings"`
	Puuid                          string     `json:"puuid" gorm:"-"`
	RetreatPings                   int        `json:"retreatPings"`
	RiotIdGameName                 string     `json:"riotIdGameName" gorm:"-"`
	RiotIdTagline                  string     `json:"riotIdTagline" gorm:"-"`
	SummonerLevel                  int        `json:"summonerLevel" gorm:"-"`
	LongestTimeSpentLiving         int        `json:"longestTimeSpentLiving"`
	MagicDamageDealt               int        `json:"magicDamageDealt"`
	TeamId                         int        `json:"teamId"`
	TeamPosition                   string     `json:"teamPosition" gorm:"index_champion_position"`
	TimeCCingOthers                int        `json:"timeCCingOthers"`
	TotalDamageDealtToChampions    int        `json:"totalDamageDealtToChampions"`
	TotalMinionsKilled             int        `json:"totalMinionsKilled"`
	TotalTimeSpentDead             int        `json:"totalTimeSpentDead"`
	TrueDamageDealtToChampions     int        `json:"trueDamageDealtToChampions"`
	VisionClearedPings             int        `json:"visionClearedPings"`
	VisionScore                    int        `json:"visionScore"`
	WardsKilled                    int        `json:"wardsKilled"`
	WardsPlaced                    int        `json:"wardsPlaced"`
	Win                            bool       `json:"win"`
}

// Challenges of the player for this match.
// Some entries like KDA and GoldPerMinute can be fetched here.
// However we can calculate at runtime without storing it.
type Challenges struct {
	AbilityUses        int `json:"abilityUses"`
	ControlWardsPlaced int `json:"controlWardsPlaced"`
	SkillshotsDodged   int `json:"skillshotsDodged"`
	SkillshotsHit      int `json:"skillshotsHit"`
}

// TeamInfo contains the bans, id and if the team won.
type TeamInfo struct {
	Bans   []Ban `json:"bans"`
	TeamId int   `json:"teamId"`
	Win    bool  `json:"win"`
}

// Ban information.
type Ban struct {
	ChampionId int `json:"championId"`
	PickTurn   int `json:"pickTurn"`
}
