package dto

// Result of a tierlist fetch.
type TierlistResult struct {
	Bancount     int
	Banrate      float64
	ChampionId   int
	Pickcount    int
	Pickrate     float64
	TeamPosition string
	Winrate      float64
}

// The fulltierlist we will return.
type FullTierlist struct {
	Champion     map[string]any // Will be a simplified version of champion.Champion, without spell data.
	BanCount     int
	Banrate      float64
	PickCount    int
	PickRate     float64
	TeamPosition string
	WinRate      float64
}
