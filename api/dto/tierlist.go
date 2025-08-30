package dto

// Result of a tierlist fetch.
type TierlistResult struct {
	BanCount     int     `json:"banCount"`
	BanRate      float64 `json:"banRate"`
	ChampionId   int     `json:"championId"`
	PickCount    int     `json:"pickCount"`
	PickRate     float64 `json:"pickRate"`
	TeamPosition string  `json:"teamPosition"`
	WinRate      float64 `json:"winRate"`
}

// The fulltierlist we will return.
type FullTierlist struct {
	Champion     map[string]any `json:"champion"` // simplified version of champion.Champion
	BanCount     int            `json:"banCount"`
	Banrate      float64        `json:"banRate"`
	PickCount    int            `json:"pickCount"`
	PickRate     float64        `json:"pickRate"`
	TeamPosition string         `json:"teamPosition"`
	WinRate      float64        `json:"winRate"`
}
